"""
Kademlia DHT implementation
"""
import json
from time import time
from random import randint
from secrets import randbits
from threading import Thread
from queue import Queue
from encodings.base64_codec import base64_encode, base64_decode
import uuid
from logging import debug, info, warning, error
from arxen_dht.networking import NetworkHandler


class UnknownResponse(Exception): pass


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
    def generate_id(bit_len: int = KadProperties.PARAM_NAMESPACE_SIZE) -> int:
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

    def get_serialized(self) -> bytes:
        # TODO for now just pass ASCII json
        return self.get_json().encode()

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
    def __init__(self, remote_node: Node, seq_value=None, *args, **kwargs):
        if not seq_value: seq_value = randint(1, 2 ** 64)
        super().__init__(remote_node=remote_node,
                         rpc_command="PING", command_arg={"seq_value": seq_value}, *args, **kwargs)


class StoreRPC(RequestRPC):
    def __init__(self, value_id: int, data: bytes, *args, **kwargs):
        super().__init__(rpc_command="STORE", command_arg={"value_id": str(value_id),
                                                           "data": str(base64_encode(data))}, *args, **kwargs)


class ResponseRPC(KadRPC):

    def __init__(self, request_rpc_uuid_b64: str, response_data_type: str, response_data, *args, **kwargs):
        """
        :param response_type: int, str, dict - how recipient should parse data
        """
        self.request_rpc_uuid_b64 = request_rpc_uuid_b64
        self.response_data_type = response_data_type
        self.response_data = response_data

        append_part = {"request_rpc_uuid": request_rpc_uuid_b64, "response_data_type": response_data_type,
                       "response_data": response_data}
        super().__init__(rpc_type="RESPONSE", append_part=append_part, *args, **kwargs)


class FindNodeResponseRPC(ResponseRPC):
    def __init__(self, nodes: list, request_rpc_uuid_b64, *args, **kwargs):
        self.nodes = nodes
        request_rpc_uuid_b64 = request_rpc_uuid_b64
        response_data_type = "node_list"
        response_data = self._fill_nodes()
        super().__init__(request_rpc_uuid_b64=request_rpc_uuid_b64, response_data_type=response_data_type,
                         response_data=response_data, *args, **kwargs)

    def _fill_nodes(self) -> list:
        resp = []
        for node in self.nodes:
            resp.append({"id": node.node_id, "ip": node.ip_info[0], "port": node.ip_info[1]})
        return resp

    @classmethod
    def from_dict(cls, data_dict: dict):
        nodes_jsoned = data_dict["response_data"]
        nodes = []
        for n in nodes_jsoned:
            nodes.append(Node(n["id"], (n["ip"], n["port"])))
        return cls(nodes=nodes, request_rpc_uuid_b64=data_dict["request_rpc_uuid"])


class FindValueResponseRPC(ResponseRPC):
    def __init__(self, data, request_rpc_uuid_b64, *args, **kwargs):
        """
        :param data: either list of Nodes OR bytes (acting as value)
        """
        self.data = data
        self.recognized_type = self._recognize_type()
        response_data = self._parse_data()
        request_rpc_uuid_b64 = request_rpc_uuid_b64
        super().__init__(request_rpc_uuid_b64=request_rpc_uuid_b64, response_data_type=self.recognized_type,
                         response_data=response_data, *args, **kwargs)

    def _recognize_type(self):
        # determine if Node list or value was passed to the constructor
        if type(self.data) is list:
            ret = "node_list"  # this node doesn't store the value so the nearest nodes were returned instead
        elif type(self.data) is bytes:
            ret = "value"
        else:
            raise TypeError("data can be type of bytes or list of nodes only")
        return ret

    def _parse_data(self):
        if self.recognized_type == "node_list":
            parsed_data = []
            for node in self.data:
                parsed_data.append({"id": node.node_id, "ip": node.ip_info[0], "port": node.ip_info[1]})
        elif self.recognized_type == "value":
            b64_value = base64_encode(self.data)
            parsed_data = b64_value
        else:
            raise TypeError("urecognized type in recognized_type")
        return parsed_data

    @classmethod
    def from_dict(cls, data: dict):
        response_data_type = data["response_data_type"]
        if response_data_type == "node_list":
            nodes_jsoned = data["response_data"]
            nodes = []
            for n in nodes_jsoned:
                nodes.append(Node(n["id"], (n["ip"], n["port"])))
            return cls(nodes, data["request_rpc_uuid"])
        elif response_data_type == "value":
            decoded_value = base64_decode(data["response_data"])
            return cls(decoded_value, data["request_rpc_uuid"])


class PingResponseRPC(ResponseRPC):
    def __init__(self, seq_value: int, request_rpc_uuid_b64, *args, **kwargs):
        request_rpc_uuid_b64 = request_rpc_uuid_b64
        response_data_type = "seq_value"
        response_data = seq_value
        super().__init__(request_rpc_uuid_b64=request_rpc_uuid_b64, response_data_type=response_data_type,
                         response_data=response_data, *args, **kwargs)

    @classmethod
    def from_dict(cls, data: dict):
        return cls(data["response_data"], data["request_rpc_uuid"])


class StoreResponseRPC(ResponseRPC):
    def __init__(self, store_successfull, request_rpc_uuid_b64, *args, **kwargs):
        """
        :param store_successfull: boolean whether store operation ended with success or not
        """
        request_rpc_uuid_b64 = request_rpc_uuid_b64
        response_data_type = "success_boolean"
        response_data = store_successfull
        super().__init__(request_rpc_uuid_b64=request_rpc_uuid_b64, response_data_type=response_data_type,
                         response_data=response_data, *args, **kwargs)

    @classmethod
    def from_dict(cls, data: dict):
        return cls(data["response_data"], data["request_rpc_uuid"])


class KadTask(Thread):
    """
    Represents generic task which is performed e.g. FINDing_NODEs requires underlying nodes interactions, so this class
    will keep needed queues and other subtasks, BASE CLASS FOR CONCURRENCY
    """

    _KadTaskList = []
    _KadTaskCounter = 0

    def __init__(self, parent=None, killed: bool = False, facility: str = "", egress_queue: Queue = None):
        super().__init__()
        self.setName("KadTask(Thread)-{}-{}".format(facility, KadTask._KadTaskCounter))
        KadTask._KadTaskCounter += 1
        KadTask._KadTaskList.append(self)

        self.children_tasks = []
        self.parentTask = parent
        self.killed = killed

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

    def send_result(self, result):
        """
        push results to egress queue
        :param results:
        :return:
        """

        self.egress_queue.put(result)

    def join_ingress_queue(self, timeout: int = 60):
        """
        simply waits for data in ingress queue and return first value
        """
        return self.ingress_queue.get(block=True, timeout=timeout)



class KadListenerTask(KadManageableTask):
    def __init__(self, remote_node=None, *args, **kwargs):
        self.remote_node = remote_node
        self.registered_tasks = []
        super().__init__(*args, **kwargs)

    def register(self, task: KadManageableTask):
        """
        RequestTasks will register in the ListenForRPCs in order to allow this listener
        to properly pass response to previous request
        """
        self.registered_tasks.append(task)

    def unregister(self, task: KadManageableTask):
        self.registered_tasks.remove(task)

    def get_registered_task_by_request_uuid(self, b64_request_uuid: str) -> KadManageableTask:
        for task in self.registered_tasks:
            if str(base64_encode(task.rpc_uuid)) == b64_request_uuid:
                return task
        raise KeyError("No registered KadRequestTask with given uuid")


class KadRequestTask(KadManageableTask):

    def __init__(self, this_node: Node, routing_table: _KadRoutingTable, response_listener: KadListenerTask,
                 remote_node: Node = None, network_handler: NetworkHandler = None, *args, **kwargs):
        self.this_node = this_node
        self.remote_node = remote_node
        self.routing_table = routing_table
        self.network_handler = network_handler
        self.response_listener = response_listener
        super().__init__(*args, **kwargs)


class ListenForRPCs(KadListenerTask):
    """
    is responsible for recognition of type of RPC: data, ping echo, FIND_NODE response etc.
    """

    def __init__(self, network_handler: NetworkHandler, *args, **kwargs):
        self.network_handler = network_handler
        super().__init__(facility="ListenForRPCs", *args, **kwargs)

    def run(self) -> None:
        datagram_rcv_buff = self.network_handler.get_rcv_buff()
        while not self.killed:
            packets = datagram_rcv_buff.join()
            for packet in packets:
                # json should appear in the packet
                try:
                    data = json.loads(packet)
                except json.JSONDecodeError:
                    warning("failed to decode jsoned packet")
                    continue  # go to the next packet

                if data["type"] == "REQUEST":
                    pass  # handle request
                elif data["type" == "RESPONSE"]:
                    try:
                        responsible_task = self.get_responsible_task(data)
                        response = self.build_response_from_data(data, responsible_task)
                        responsible_task.ingress_queue.put(response)
                    except UnknownResponse as e:
                        warning(e)
                        continue  # GOTO next rpc

    @staticmethod
    def build_response_from_data(rpc_data: dict, responsible_task: KadManageableTask) -> ResponseRPC:

        response_type = type(responsible_task)
        if response_type is FindNodeRPC:
            return FindNodeResponseRPC.from_dict(rpc_data)
        elif response_type is FindValueRPC:
            return FindValueResponseRPC.from_dict(rpc_data)
        elif response_type is StoreRPC:
            return StoreResponseRPC.from_dict(rpc_data)
        elif response_type is PingRPC:
            return PingResponseRPC.from_dict(rpc_data)

    def get_responsible_task(self, rpc_data: dict):
        request_b64_uuid = rpc_data["request_rpc_uuid"]
        try:
            responsible_request_task = self.get_registered_task_by_request_uuid(request_b64_uuid)
        except KeyError:
            raise UnknownResponse(
                "Response with unregistered request_uuid was received, uuid={}".format(request_b64_uuid))
        return responsible_request_task


class FindNodeTask(KadRequestTask):
    def __init__(self, lookup_node_id: int, entry_point: bool = False, *args, **kwargs):
        """
        :param node_id: node we are trying to find
        :param EntryPoint: Flag is set true if this task is used to initialize performing lookup process
        """
        self.lookup_node = lookup_node_id
        self.entry_point = entry_point
        super().__init__(facility="FindNodeTask", *args, **kwargs)

    def run(self):
        debug("running {}".format(self))
        nearest_nodes = []
        queried_nodes = []

        # Run as entry point
        if self.entry_point:
            # First select my nearest nodes
            my_nearest_nodes = self.routing_table.route_to(self.lookup_node)
            nearest_nodes.extend(my_nearest_nodes)
            for node in my_nearest_nodes:
                self.spawn_task(FindNodeTask(self.lookup_node, remote_node=node))
                # Here handle returned nodes from childs

        # Run as a child querying a remote node
        if self.remote_node:
            new_rpc = FindNodeRPC(self.lookup_node, self.remote_node, node=self.this_node)
            self.network_handler.send_bytes(new_rpc.get_serialized(), self.remote_node.ip_info)
            self.response_listener.register(self)

            try:
                response: FindNodeResponseRPC = self.join_ingress_queue()   # wait for Response() from ListenFromRPCs()
                debug("{}: received response {}".format(self, response))
            except TimeoutError:
                # requested node hasn't responded
                # TODO remove this inactive node from the routing table
                self.response_listener.unregister(self)  # DO NOT wait for the answer anymore
                return  # terminate this task
            for node in response.nodes:
                self.send_result(node)
            # Successfully passed nodes to the parent so now die
            self.response_listener.unregister(self)
            return


class KadEngine:
    def __init__(self):
        # preparing network
        self.network_handler = NetworkHandler()

        self.my_node = Node(Node.generate_id(), ("::", 55667))
        self.routing_table = _KadRoutingTable(self.my_node.node_id)

        self.rpc_listener = ListenForRPCs(self.network_handler)

    @staticmethod
    def distance(self, node_a: Node, node_b: Node):
        """
        :return: XOR distance between a and b
        """
        return node_a.node_id ^ node_b.node_id
