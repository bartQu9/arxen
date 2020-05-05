"""
Kademlia DHT implementation
"""
import json
from time import time
from secrets import randbits
from threading import Thread
from queue import Queue
from encodings.base64_codec import base64_encode
import uuid
from logging import debug, info, warning, error


class KadProperties:
    """
    Simply stores Kademlia properties as alpha or k params
    """
    PARAM_ALPHA = 10
    PARAM_K = 20
    PARAM_NAMESPACE_SIZE = 256


class Node:
    """
    Represents a Kademlia node
    """

    def __init__(self, node_id: int, ip_info: tuple):
        """
        :param node_id: Kademlia node ID
        :param ip_info: IP,port tuple
        """
        self.node_id = node_id
        self.ip_info = ip_info

        return

    @staticmethod
    def _generate_id(bit_len: int = KadProperties.PARAM_NAMESPACE_SIZE) -> int:
        return randbits(bit_len)


class _KadRoutingTable:
    def __init__(self, my_node_id: int, bucket_size: int = KadProperties.PARAM_K,
                 buckets_count: int = KadProperties.PARAM_NAMESPACE_SIZE, ):
        self.my_node_id = my_node_id
        self.bucket_size = bucket_size
        self.buckets_count = buckets_count
        self.buckets = self._create_buckets()

    def _create_buckets(self) -> tuple:
        return tuple([[] for _ in range(self.buckets_count)])

    def _change_on_which_bit(self, comparable_id: int) -> int:
        """
        Returns on nth bit difference e.g:
        ID1: 0011
        ID2: 0111
              ^------returns 2 (second bit counting from 0 right)
        """
        xor = self.my_node_id ^ comparable_id
        nth_position = 0
        while xor != 1:
            xor >>= 1
            nth_position += 1
        return nth_position

    def insert_node(self, node: Node):
        """
        stores a node in an appropriate bucket
        K-buckets: [1;2) [2;4) [4;8) [8;16) ..... // [distance]
               ==  [1]   [2;3] [4;7] [8;15] .....
        """
        which_bucket = self._change_on_which_bit(node.node_id)
        self.buckets[which_bucket].append(Node)

    # TODO append_node() which removes exceeding dead nodes

    def get_bucket(self, index: int) -> list:
        return self.buckets[index]

    def route_to(self, remote_node_id: int, next_hops_count=KadProperties.PARAM_ALPHA) -> list:
        """
        Returns the closest nodes we already know; By default alpha nodes
        This is a key functionality needed for performing FIND_NODE RPC
        """
        collected_nodes = list()
        # If we're lucky we'll get all nodes from this single bucket
        # otherwise we'll have to collect missing nodes from "further" buckets
        next_closest_bucket = self._change_on_which_bit(remote_node_id)

        while len(collected_nodes) < next_hops_count:
            collected_nodes += self.get_bucket(next_closest_bucket)
            next_closest_bucket += 1
            if next_closest_bucket > (self.buckets_count - 1):
                break

        return collected_nodes[:next_hops_count]


class KadRPC:
    def __init__(self, node: Node, rpc_type: str, append_part: dict = None):
        # TODO add a DEFAULT_NODE which is set for this local node to prevent passing Node arg in subsequent RPCs calls
        """
        :param rpc_type: REQUEST or RESPONSE
        :param node: Node which interface will send this RPC (req or resp)
        """
        self.node = node
        self.rpc_type = rpc_type
        self.append_part = append_part
        self.rpc_uuid = uuid.uuid1()

        self.dict_repr = self.to_dict_representation()

    def to_dict_representation(self):
        sending_node = {"id": self.node.node_id,
                        "ip": self.node.ip_info[0],
                        "port": self.node.ip_info[1]}
        rpc = {"node": sending_node}
        rpc.update({"type": self.rpc_type})

        # append RPC parts from children class
        if self.append_part:
            rpc.update(self.append_part)

        rpc_uuid_b64 = base64_encode(self.rpc_uuid.bytes)[0].decode()
        rpc.update({"rpc_uuid": rpc_uuid_b64})

        return rpc

    def get_json(self):
        return json.dumps(self.to_dict_representation())

    def __repr__(self):
        return "{}: {}".format(self.__class__.__name__, self.dict_repr)


class RequestRPC(KadRPC):

    def __init__(self, remote_node: Node, rpc_command: str, command_arg: dict, *args, **kwargs):
        self.rpc_command = rpc_command
        self.command_arg = command_arg
        self.remote_node = remote_node
        super().__init__(rpc_type="REQUEST", append_part={"command": rpc_command, "arg": command_arg}, *args, **kwargs)


class FindNodeRPC(RequestRPC):
    def __init__(self, lookup_node_id: int, remote_node: Node, *args, **kwargs):
        """
        lookup_node_id: node we're looking for
        """
        super().__init__(rpc_command="FIND_NODE", command_arg={"node_id": str(lookup_node_id)},
                         remote_node=remote_node, *args, **kwargs)


class FindValueRPC(RequestRPC):
    def __init__(self, value_id: int, *args, **kwargs):
        """
        value_id: value ID of data we're trying to GET
        """
        super().__init__(rpc_command="FIND_VALUE", command_arg={"value_id": str(value_id)}, *args, **kwargs)


class PingRPC(RequestRPC):
    def __init__(self, node_id: int, *args, **kwargs):
        super().__init__(rpc_command="PING", command_arg={"node_id": str(node_id)}, *args, **kwargs)


class StoreRPC(RequestRPC):
    def __init__(self, value_id: int, data: bytes, *args, **kwargs):
        super().__init__(rpc_command="STORE", command_arg={"value_id": str(value_id),
                                                           "data": str(base64_encode(data))}, *args, **kwargs)


class ResponseRPC(KadRPC):

    def __init__(self, request_rpc_uuid: uuid.UUID, response_data, *args, **kwargs):
        """
        :param response_type: int, str, dict - how recipient should parse data
        """
        self.request_rpc_uuid = request_rpc_uuid
        self.response_data = response_data

        append_part = {"request_rpc_uuid": request_rpc_uuid, "response_data": response_data}
        super().__init__(rpc_type="RESPONSE", append_part=append_part, *args, **kwargs)


class KadTask(Thread):
    """
    Represents generic task which is performed e.g. FINDing_NODEs requires underlying nodes interactions, so this class
    will keep needed queues and other subtasks, BASE CLASS FOR CONCURRENCY
    """

    _KadTaskList = []
    _KadTaskCounter = 0

    def __init__(self, this_node: Node, routing_table: _KadRoutingTable,
                 parent=None, facility: str = "", egress_queue: Queue = None):
        super().__init__()
        self.setName("KadTask(Thread)-{}-{}".format(facility, KadTask._KadTaskCounter))
        KadTask._KadTaskCounter += 1
        KadTask._KadTaskList.append(self)
        self.this_node = this_node
        self.routing_table = routing_table

        self.children_tasks = []
        self.parentTask = parent

        self.ingress_queue = Queue()
        self.egress_queue = egress_queue

        debug("Initialized {}".format(self.getName()))

    @staticmethod
    def get_existing_kad_tasks() -> list:
        return KadTask._KadTaskList

    def __del__(self):
        debug("Destroying {}".format(self))
        KadTask._KadTaskList.remove(self)
        KadTask._KadTaskCounter -= 1


class KadManageableTask(KadTask):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def spawn_task(self, task: KadTask):
        debug("{} spawns {}".format(self, task))
        # attaching queue to allow pass children results to its parent
        if task.egress_queue is None:
            task.egress_queue = self.ingress_queue

        self.children_tasks.append(task)
        task.start()

    def send_results(self, results):
        """
        push results to egress queue
        :param results:
        :return:
        """
        if results.__iter__:
            for res in results:
                self.egress_queue.put(res)
        else:
            self.egress_queue.put(results)

    def join_queue(self, timeout: int = 60):
        """
        simply waits for data in ingress queue and return first value
        """
        self.ingress_queue.get(block=True, timeout=timeout)


class FindNodeTask(KadManageableTask):
    def __init__(self, node_id: int, entry_point: bool = False, *args, **kwargs):
        """
        :param node_id: node we are trying to find
        :param EntryPoint: Flag is set true if this task is used to initialize performing lookup process
        """
        self.lookup_node = node_id
        self.entry_point = entry_point
        super().__init__(facility="FindNodeTask", *args, **kwargs)

    def run(self):
        debug("running {}".format(self))
        nearest_nodes = []
        queried_nodes = []

        if self.entry_point:
            # First select my nearest nodes
            my_nearest_nodes = self.routing_table.route_to(self.lookup_node)
            for node in my_nearest_nodes:
                self.spawn_task(FindNodeTask(self.lookup_node))




class KadEngine:
    def __init__(self):
        pass

    @staticmethod
    def distance(self, node_a: Node, node_b: Node):
        """
        :return: XOR distance between a and b
        """
        return node_a.node_id ^ node_b.node_id
