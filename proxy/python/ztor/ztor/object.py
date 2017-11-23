from . import proxy_pb2
from . import proxy_pb2_grpc


class Object():
    def __init__(self, channel):
        self.stub = proxy_pb2_grpc.ObjectServiceStub(channel)

    def write(self, key, value, ref_list=None,
              prev_key=None, prev_meta=None, meta=None):
        req = proxy_pb2.WriteRequest(
            key=key,
            meta=meta,
            value=value,
            referenceList=ref_list,
            prevKey=prev_key,
            prevMeta=prev_meta
        )
        return self.stub.Write(req)

    def write_file(self, key, file_path, ref_list=None,
                   prev_key=None, prev_meta=None, meta=None):

        req = proxy_pb2.WriteFileRequest(
            key=key,
            meta=meta,
            filePath=file_path,
            referenceList=ref_list,
            prevKey=prev_key,
            prevMeta=prev_meta
        )
        return self.stub.WriteFile(req)

    def write_stream_iterator(self, key, it, ref_list=None,
                              prev_key=None, prev_meta=None, meta=None):

        for val in it:
            req = proxy_pb2.WriteStreamRequest(
                key=key,
                meta=meta,
                value=val,
                referenceList=ref_list,
                prevKey=prev_key,
                prevMeta=prev_meta
            )
            yield req

    def write_stream(self, key, it, ref_list=None,
                     prev_key=None, prev_meta=None, meta=None):

        rpc_it = self.write_stream_iterator(key, it, ref_list, prev_key,
                                            prev_meta, meta)
        return self.stub.WriteStream(rpc_it)

    def read(self, key, meta=None):
        req = proxy_pb2.ReadRequest(
            key=key,
            meta=meta
        )
        return self.stub.Read(req)

    def read_file(self, key, file_path, meta=None):
        req = proxy_pb2.ReadFileRequest(
            key=key,
            meta=meta,
            filePath=file_path,
        )
        return self.stub.ReadFile(req)

    def read_stream(self, key, meta=None):
        req = proxy_pb2.ReadRequest(
            key=key,
            meta=meta
        )

        stream = self.stub.ReadStream(req)

        for s in stream:
            yield s

    def delete(self, key, meta=None):
        req = proxy_pb2.DeleteRequest(
            key=key,
            meta=meta
        )
        return self.stub.Delete(req)

    def walk(self, start_key, from_epoch, to_epoch):
        req = proxy_pb2.WalkRequest(
            startKey=start_key,
            fromEpoch=int(from_epoch) * 1000 * 1000 * 1000,
            toEpoch=int(to_epoch) * 1000 * 1000 * 1000
        )

        objects = self.stub.Walk(req)

        for obj in objects:
            yield obj

    def append_reference_list(self, key, ref_list, meta=None):
        req = proxy_pb2.ReferenceListRequest(
            key=key,
            referenceList=ref_list,
            meta=meta
        )

        return self.stub.AppendReferenceList(req)

    def remove_reference_list(self, key, ref_list, meta=None):
        req = proxy_pb2.ReferenceListRequest(
            key=key,
            referenceList=ref_list,
            meta=meta
        )

        return self.stub.RemoveReferenceList(req)

    def check(self, key):
        req = proxy_pb2.CheckRequest(
            key=key
        )

        return self.stub.Check(req)
        # print(resp)
        # assert(proxy_pb2.CheckReply.ok == resp.status)
