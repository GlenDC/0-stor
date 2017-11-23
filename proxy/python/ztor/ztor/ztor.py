import grpc

from .object import Object
from .namespace import Namespace


class ztor():
    def __init__(self, addr):
        channel = grpc.insecure_channel(addr)
        self.object = Object(channel)
        self.namespace = Namespace(channel)
