import socket
import threading
import queue

from logging import debug, info, warning, error


class NetworkingException(Exception):
    pass


class NetworkHandlerException(NetworkingException):
    pass


class BindingException(NetworkHandlerException):
    pass


class SockHandler:
    def __init__(self):

        self.rcv_buff = queue.Queue()
        self.send_buff = queue.Queue()
        self.sock = self.bind_sock()
        self.child_threads = []

        self.start_listening_sending()

    def bind_sock(self, port=55667, addr_fam=socket.AF_INET6):
        s = socket.socket(addr_fam, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
        try:
            s.bind(("::", port))
            s.settimeout(1)
        except OSError as err:
            debug("{} failed to bind: {}".format(self, err))
            raise BindingException
        return s

    def _listen(self, buff_size=1460):
        """
        blocking!
        :param buff_size: maximum datagram size
        :param rcv_queue: supposed to be a FIFO queue
        """
        debug("starting listening, {}".format(threading.current_thread().getName()))
        while self.sock.fileno() > 0:  # check whether socket.close()
            try:
                self.rcv_buff.put(self.sock.recvmsg(buff_size))
            except socket.timeout:
                pass

    def _send(self):
        """
        blocking
        """
        debug("starting sending, {}".format(threading.current_thread().getName()))
        while self.sock.fileno() > 1:
            try:
                data, addr = self.send_buff.get(timeout=1)
                self.sock.sendto(data, addr)
            except (queue.Empty, socket.timeout):
                pass

    def send(self, data, addr):
        # should not block
        self.send_buff.put((data, addr))

    def receive(self, count=0):
        if count == 0:
            count = self.rcv_buff.qsize()
        received = []
        [received.append(self.rcv_buff.get()) for _ in range(count)]
        return received

    def close_socket(self):
        self.sock.close()
        for t in self.child_threads:
            t.join()
        # clear data left in queues
        self.rcv_buff = queue.Queue()
        self.send_buff = queue.Queue()

    def start_listening_sending(self):
        self.child_threads.append(threading.Thread(target=self._listen, name="listening_thread"))
        self.child_threads.append(threading.Thread(target=self._send, name="sending_thread"))
        for t in self.child_threads:
            t.start()

    def __del__(self):
        self.close_socket()



class NetworkHandler:
    class Task:
        def __init__(self, task_type, in_data, out_data, flags=None):
            if flags is None:
                flags = []
            self.type = task_type
            self.in_data = in_data
            self.out_data = out_data

    def __init__(self):
        pass

    def run(self):
        # get tasks from task queue e.g. send()..
        pass

