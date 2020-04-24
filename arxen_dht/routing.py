"""
Kademlia DHT implementation
"""

from secrets import randbits


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
    def _generate_id(self, bit_len: int = KadProperties.PARAM_NAMESPACE_SIZE) -> int:
        return randbits(bit_len)


class _KadRoutingTable:
    def __init__(self, bucket_size: int = KadProperties.PARAM_K,
                 buckets_count: int = KadProperties.PARAM_NAMESPACE_SIZE):
        self.bucket_size = bucket_size
        self.buckets_count = buckets_count
        self.buckets = self.create_buckets()

    def create_buckets(self) -> tuple:
        return tuple([[] for _ in range(self.buckets_count)])

    def insert(self, node: Node):
        """
        stores a node in an appropriate bucket
        """


class KadEngine:
    def __init__(self):
        pass

    @staticmethod
    def distance(self, node_a: Node, node_b: Node):
        """
        :return: XOR distance between a and b
        """
        return node_a.node_id ^ node_b.node_id
