from ztor import ztor


def test_object_lib(zt):
    print("--- write value and read it back--- ")
    my_key = bytes("my_key", 'utf-8')

    resp = zt.object.walk(my_key, bytes("my value", 'utf-8'))
    print(resp)

    resp = zt.object.read(my_key)
    print(resp)

    print('--- write file and read it back')
    file_key = bytes("file-key", 'utf-8')
    file_to_upload = '/tmp/zerostorserver'

    resp = zt.object.write_file(file_key, file_to_upload)
    print(resp.meta.key)

    resp = zt.object.read_file(file_key, "/tmp/downloaded")
    print(resp.referenceList)

    print(' -- delete value and confirm deleted--')
    resp = zt.object.delete(my_key)
    print(resp)

    try:
        zt.object.read(my_key)
    except Exception:
        print("=> OK")

    print('-- delete file and confirm deleted --')
    resp = zt.object.delete(file_key)
    print(resp)

    try:
        zt.object.read_file(file_key, '/tmp/downloaded')
    except Exception:
        print('--> OK')


def test_namespace(zt):
    res = zt.namespace.create_jwt("your_namespace", read=True, write=True)
    print(res)


if __name__ == '__main__':
    addr = '127.0.0.1:12121'

    zt = ztor(addr)
    test_object_lib(zt)
    test_namespace(zt)
