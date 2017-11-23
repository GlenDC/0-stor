from .ztor import ztor

from . import proxy_pb2

CheckStatusOK = proxy_pb2.CheckReply.ok
CheckStatusCorrupted = proxy_pb2.CheckReply.missing
CheckStatusMissing = proxy_pb2.CheckReply.corrupted