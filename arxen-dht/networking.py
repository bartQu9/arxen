import socket
import threading
import queue
from logging import debug, info, warning, error

class NetworkingException(Exception):
    pass
class HostListerException(NetworkingException):
    pass
class BindingException(HostListerException):
    pass

class NetworkHandler:

    def request(self, host, rpc):
        pass


class HostListener(threading):

    class Task:
        def __init__(self, type, in_data, out_data, flags=None):
            if flags is None:
                flags = []
            self.type = type
            self.in_data = in_data
            self.out_data = out_data

    def __init__(self, sock=None):
        super().__init__()
        self.sock = sock

        self.bind_sock()

    def bind_sock(self, port=55667, addr_fam=socket.AF_INET6):
        s = self.sock
        s = socket.socket(addr_fam, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
        try:
            s.bind(("::", port))
        except OSError as err:
            debug("{} failed to bind: {}".format(self, err))
            raise BindingException


    def listen(self, rcv, buff_size=1460):
        """
        blocking!
        :param buff_size: maximum datagram size
        :param rcv: supposed to be a FIFO queue
        """
        while self.sock:
            rcv.put(self.sock.recvmsg(buff_size))

    def run(self, task_queue):
        #get tasks from task queue e.g. send()..

