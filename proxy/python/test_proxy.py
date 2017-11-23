import unittest
import subprocess
import os
import tempfile
import random
import string

import time
import yaml
import delegator

from ztor.ztor import ztor
from ztor.ztor import CheckStatusOK, CheckStatusCorrupted

PROXY_ADDR = '127.0.0.1:12121'
DELETER_PROXY_ADDR = '127.0.0.1:12122'
BLOCK_SIZE = 4096

class ProxyTest(unittest.TestCase):
    """
    ProxyTest is class to test 0-stor proxy
    """
    @classmethod
    def setUpClass(self):
        config_file = os.environ['GOPATH'] + '/src/github.com/zero-os/0-stor/proxy/python/config.yaml'
        self.cluster = ZtorCluster(config_file)
        self.cluster.start()
        self.zt = ztor(PROXY_ADDR)

    @classmethod
    def tearDownClass(self):
        self.cluster.stop()
        self.cluster.cleanup()

    def setUp(self):
        # self.cluster.start()
        pass

    def tearDown(self):
        # self.cluster.stop()
        pass

    def randomword(self, length):
        return ''.join(random.choice(string.ascii_letters + string.digits) for i in range(length))

    def test_write_read_delete(self):
        """ test object write, read, and delete"""
        print('test_write_read_delete')

        kvs = []

        # initialize keys & values
        for i in range(0, 50):
            key = bytes('write_read_delete_my_{}_key'.format(i), 'utf-8')
            val = bytes(self.randomword(4096), 'utf-8')
            kvs.append({'key': key, 'val': val})

        # write values
        for kv in kvs:
            self.zt.object.write(kv['key'], kv['val'])

        # read values
        for kv in kvs:
            resp = self.zt.object.read(kv['key'])
            self.assertEqual(resp.value, kv['val'])

        # delete values
        for kv in kvs:
            self.zt.object.delete(kv['key'])

        # make sure that read failed for us this time
        for kv in kvs:
            with self.assertRaises(Exception):
                self.zt.object.read(kv['key'])

    def test_file_write_read_delete(self):
        """ test object write_file, read_file, and delete"""
        print('test_file_write_read_delete')

        key = bytes('file_write_read_delete_key', 'utf-8')

        # generate random file
        f = tempfile.NamedTemporaryFile()
        with open(f.name, 'w') as out:
            for x in range(1, 1024):
                out.write(self.randomword(1024))

        # upload file
        self.zt.object.write_file(key, f.name)

        # download file to temporary file
        downloaded = tempfile.NamedTemporaryFile()
        self.zt.object.read_file(key, downloaded.name)

        self.assertEqual(open(f.name).read(), open(downloaded.name).read())

        # delete file
        self.zt.object.delete(key)

        # make sure read file got exception
        with self.assertRaises(Exception):
            self.zt.object.read_file(key)

    def _stream_generator(self, val, block_size=BLOCK_SIZE):
        while len(val) > 0:
            chunk = val[:block_size]
            val = val[block_size:]
            yield bytes(chunk, 'utf-8')

    def test_write_stream(self):
        # test write stream to proxy
        print('test_write_stream')

        key = bytes('write_stream_key', 'utf-8')

        # write the stream
        value = self.randomword(BLOCK_SIZE * 10)
        it = self._stream_generator(value)
        self.zt.object.write_stream(key, it)

        # make sure we have valid data
        resp = self.zt.object.read(key)
        self.assertEqual(bytes(value, 'utf-8'), resp.value)

    def test_read_stream(self):
        """ test_read_stream
        test steps:
            - upload value
            - read value as stream
        """
        print('test_read_stream')

        # upload the value
        key = bytes('read_stream_key', 'utf-8')
        val = bytes(self.randomword(BLOCK_SIZE * 10), 'utf-8')

        self.zt.object.write(key, val)

        # read value as stream
        received = bytes(0)
        stream = self.zt.object.read_stream(key)

        for s in stream:
            received += s.value

        self.assertEqual(val, received)

    def test_walk(self):
        # test write file and walk over it
        print('test walk')

        kvs = []

        # initialize keys & values
        for i in range(0, 50):
            key = bytes('walk_{}_key'.format(i), 'utf-8')
            val = bytes(self.randomword(BLOCK_SIZE), 'utf-8')
            kvs.append({'key': key, 'val': val})

        # write values
        start_epoch = time.time()
        prev_key = None
        for kv in kvs:
            self.zt.object.write(kv['key'], kv['val'], prev_key=prev_key)
            prev_key = kv['key']

        end_epoch = time.time()

        # walk over it
        objects = self.zt.object.walk(kvs[0]['key'], start_epoch, end_epoch)

        idx = 0
        for obj in objects:
            self.assertEqual(kvs[idx]['val'], obj.value)
            idx = idx + 1

    def test_append_reference_list(self):
        # test append reference list
        print('test append reference list')

        key = bytes('append_ref_key', 'utf-8')
        val = bytes(self.randomword(BLOCK_SIZE), 'utf-8')

        # write
        self.zt.object.write(key, val)

        # make sure we have no reference list
        resp = self.zt.object.read(key)
        self.assertEqual([], resp.referenceList)

        # append ref list
        ref_list = ['1234']
        self.zt.object.append_reference_list(key, ref_list)

        # make sure we have ref list
        resp = self.zt.object.read(key)
        self.assertEqual(ref_list, resp.referenceList)

    def test_remove_reference_list(self):
        print('test remove reference list')

        key = bytes('remove_ref_key', 'utf-8')
        val = bytes(self.randomword(BLOCK_SIZE), 'utf-8')
        ref_list = ['12345']

        # write it
        self.zt.object.write(key, val, ref_list=ref_list)

        # verify that we have the given ref_list
        resp = self.zt.object.read(key)
        self.assertEqual(ref_list, resp.referenceList)

        # remove it
        self.zt.object.remove_reference_list(key, ref_list)

        # verify that the reflist is not exist anymore
        resp = self.zt.object.read(key)
        # TODO enable this : self.assertEqual([], resp.referenceList)

    def test_check(self):
        print('test check')

        key = bytes('check_key_' + self.randomword(10), 'utf-8')
        val = bytes(self.randomword(BLOCK_SIZE), 'utf-8')

        # write it
        self.zt.object.write(key, val)

        # check it
        resp = self.zt.object.check(key)
        self.assertEqual(CheckStatusOK, resp.status)

        # TODO : simulate corrupted data

    @unittest.skip('skip it because we cant simulate to corrupts data')
    def test_repair(self):
        pass

    def corrupt_data(self, key):
        # TODO : we can
        dp = DeleterProxy(self.cluster.bin_dir, self.cluster.proxy.config_file,
                          DELETER_PROXY_ADDR)
        dp.start()
        zt = ztor(dp.proxy_addr)
        zt.object.write(key, bytes(1))
        dp.stop()


class ZtorCluster:
    """
    ZtorCluster wraps all 0-stor cluster
    start and stop
    """

    def __init__(self, config_file):
        self.config = yaml.load(open(config_file, 'r'))
        self.config['organization'] = os.environ['IYO_ORGANIZATION']
        self.config['iyo_app_id'] = os.environ['IYO_APP_ID']
        self.config['iyo_app_secret'] = os.environ['IYO_APP_SECRET']

        self.bin_dir = os.environ['GOPATH'] + '/src/github.com/zero-os/0-stor/bin'

    def cleanup(self):
        """ do necessary cleanup """
        pass

    def compile(self):
        """ compile all 0-stor binaries"""
        # TODO : make it smarter because we loss time for compiling
        print('compiling all binaries')
        c = delegator.run('cd ../../; make server proxy', block=True)
        print(c.out)

    def start(self):
        self.start_etcd()
        self.start_ztorserver()
        self.start_proxy()

        # TODO better way to wait than sleeping here
        print('waiting for 0-stor cluster to be fully started')
        time.sleep(10)
        print('0-stor cluster is (hopefully) fully started')

    def stop(self):
        self.stop_etcd()
        self.stop_ztorserver()
        self.stop_proxy()

    def start_etcd(self):
        self._etcd_dir = tempfile.TemporaryDirectory()
        self._etcd = subprocess.Popen(['etcd'], cwd=self._etcd_dir.name)

        # TODO don't rewrite this meta shards config
        self.config['meta_shards'] = ['127.0.0.1:2379']

    def stop_etcd(self):
        self._etcd.kill()
        self._etcd_dir.cleanup()

    def start_proxy(self):
        self.proxy_config_file = tempfile.NamedTemporaryFile(delete=True)
        with open(self.proxy_config_file.name, 'w') as conf_file:
            yaml.dump(self.config, conf_file)

        self.proxy = Proxy(self.bin_dir, self.proxy_config_file.name,
                           PROXY_ADDR)
        self.proxy.start()

    def stop_proxy(self):
        self.proxy.stop()
        self.proxy_config_file.close()

    def start_ztorserver(self):
        self.ztorserver_dirs = []
        self.ztorservers = []
        for addr in self.config['data_shards']:
            sdir = tempfile.TemporaryDirectory()
            self.ztorserver_dirs.append(sdir)

            dirname = sdir.name
            print('start zerostorserver on dir', dirname)
            serv = subprocess.Popen([self.bin_dir + '/zerostorserver',
                                    '-b', addr,
                                     '--data', '{}/data'.format(dirname),
                                     '--meta', '{}/meta'.format(dirname)],
                                    cwd=dirname)
            self.ztorservers.append(serv)

    def stop_ztorserver(self):
        for serv in self.ztorservers:
            serv.kill()

        for sdir in self.ztorserver_dirs:
            sdir.cleanup()


class Proxy:
    # Proxy is wrapper for 0-stor proxy

    def __init__(self, bin_dir, config_file, proxy_addr):
        self.bin_dir = bin_dir
        self.config_file = config_file
        self.proxy_addr = proxy_addr

    def start(self):
        self._proxy = subprocess.Popen([self.bin_dir + '/ztorproxy',
                                       '--config', self.config_file,
                                        '--address', self.proxy_addr])

    def stop(self):
        self._proxy.kill()


class DeleterProxy(Proxy):
    # DeleterProxy is proxy that only used to corrupt uploaded data

    def __init__(self, bin_dir, non_deleter_conf_file, proxy_addr):
        conf = yaml.load(open(non_deleter_conf_file, 'r'))

        # modify the proxy config
        if conf['replication_nr'] != 0:
            conf['replication_nr'] -= 1
        else:
            conf['replication_nr'] = conf['distribution_parity']

        # write our conf file
        conf_file = tempfile.NamedTemporaryFile(delete=False)
        with open(conf_file.name, 'w') as f:
            yaml.dump(conf, f)

        super().__init__(bin_dir, conf_file.name, proxy_addr)

        self.conf_file_obj = conf_file

    def start(self):
        super().start()
        time.sleep(5)

    def stop(self):
        super().stop()
        self.conf_file_obj.close()

