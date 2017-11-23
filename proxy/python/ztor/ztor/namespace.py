from . import proxy_pb2
from . import proxy_pb2_grpc


class Namespace():
    def __init__(self, channel):
        self.stub = proxy_pb2_grpc.NamespaceServiceStub(channel)

    def create_jwt(self, namespace, read=False, write=False, delete=False,
                   admin=False):
        perm = proxy_pb2.Permission(
            read=read,
            write=write,
            delete=delete,
            admin=admin
        )
        req = proxy_pb2.CreateJWTRequest(
            namespace=namespace,
            permission=perm
        )
        return self.stub.CreateJWT(req)

    def create_namespace(self, namespace):
        req = proxy_pb2.NamespaceRequest(
            namespace=namespace
        )
        return self.stub.CreateNamespace(req)

    def delete_namespace(self, namespace):
        req = proxy_pb2.NamespaceRequest(
            namespace=namespace
        )
        return self.stub.DeleteNamespace(req)

    def give_permission(self, namespace, user_id, read=False, write=False,
                        delete=False, admin=False):
        perm = proxy_pb2.Permission(
            read=read,
            write=write,
            delete=delete,
            admin=admin
        )
        req = proxy_pb2.EditPermissionRequest(
            namespace=namespace,
            userID=user_id,
            permission=perm
        )
        return self.stub.GivePermission(req)

    def remove_permission(self, namespace, user_id, read=False, write=False,
                          delete=False, admin=False):
        perm = proxy_pb2.Permission(
            read=read,
            write=write,
            delete=delete,
            admin=admin
        )
        req = proxy_pb2.EditPermissionRequest(
            namespace=namespace,
            userID=user_id,
            permission=perm
        )
        return self.stub.RemovePermission(req)

    def get_permission(self, namespace, user_id):
        req = proxy_pb2.GetPermissionRequest(
            namespace=namespace,
            userID=user_id
        )
        return self.stub.GetPermission(req)
