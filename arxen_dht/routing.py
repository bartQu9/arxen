"""
Kademlia DHT implementation
"""
import json
from secrets import randbits
from threading import Thread
from queue import Queue


class KadProperties:
    """
    Simply stores Kademlia properties as alpha or k params
    """
    PARAM_ALPHA = 20
    PARAM_K = 10
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
    def __init__(self, initiator: Node, rpc_command: str, command_arg: dict):
        self.initiator = initiator
        self.rpc_command = rpc_command
        self.command_arg = command_arg

    def to_dict_representation(self):
        requesting_node = {"id":   self.initiator.node_id,
                           "ip":   self.initiator.ip_info[0],
                           "port": self.initiator.ip_info[1]}
        command = self.rpc_command
        arg = self.command_arg

        rpc = {"node": requesting_node,
               "command": command,
               "arg": arg}
        return rpc

    def get_json(self):
        return json.dumps(self.to_dict_representation())


class FindNodeRPC(KadRPC):
    def __init__(self, node_id: int, initiator: Node):
        """
        node_id: node we're looking for
        """
        super().__init__(initiator=initiator, rpc_command="FIND_NODE", command_arg={"node_id": str(node_id)})

class FindValueRPC(KadRPC):
    def __init__(self, value_id: int, initiator: Node):
        """
        value_id: value ID of data we're trying to GET
        """
        super().__init__(initiator=initiator, rpc_command="FIND_VALUE", command_arg={"value_id": str(value_id)})
class KadRCPResponse:
    pass

class KadTask(Thread):
    """
    Represents generic task which is performed e.g. FINDing_NODEs requires underlying nodes interactions, so this class
    will keep needed queues and other subtasks, BASE CLASS FOR CONCURRENCY
    """

    _KadTaskList = []
    _KadTaskCounter = 0

    def __init__(self, parent=None, facility: str = ""):
        super().__init__()
        self.setName("KadTask(Thread)-{}-{}".format(facility, KadTask._KadTaskCounter))
        KadTask._KadTaskCounter += 1
        KadTask._KadTaskList.append(self)

        self.children_tasks = []
        self.parentTask = parent

        self.ingress_queue = Queue()
        self.egress_queue = Queue()

    @staticmethod
    def get_existing_kad_tasks() -> list:
        return KadTask._KadTaskList


class FindNodeTask(KadTask):
    def __init__(self, node_id: int, *args, **kwargs):
        """
        :param node_id: node we are trying to find
        """
        super().__init__(facility="FindNodeTask")

    def run(self):






class KadEngine:
    def __init__(self):
        pass

    @staticmethod
    def distance(self, node_a: Node, node_b: Node):
        """
        :return: XOR distance between a and b
        """
        return node_a.node_id ^ node_b.node_id
