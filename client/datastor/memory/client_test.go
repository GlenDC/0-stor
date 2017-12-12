package memory

import (
	"context"
	"crypto/rand"
	"fmt"
	mathRand "math/rand"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
	"github.com/zero-os/0-stor/client/datastor"
)

func TestNewclient(t *testing.T) {
	client := NewClient()
	require.NotNil(t, client)
}

func TestClientSetObjectErrors(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	err := client.SetObject(datastor.Object{})
	require.Equal(errNilKey, err)

	err = client.SetObject(datastor.Object{Key: []byte("foo")})
	require.Equal(errNilData, err)
}

func TestClientSetGetObject(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	obj, err := client.GetObject(nil)
	require.Equal(errNilKey, err)
	require.Nil(obj)

	obj, err = client.GetObject([]byte("foo"))
	require.Equal(datastor.ErrKeyNotFound, err)
	require.Nil(obj)

	err = client.SetObject(datastor.Object{Key: []byte("foo"), Data: []byte("bar")})
	require.NoError(err)

	obj, err = client.GetObject([]byte("foo"))
	require.NoError(err)
	require.NotNil(obj)
	require.Equal([]byte("foo"), obj.Key)
	require.Equal([]byte("bar"), obj.Data)
	require.Empty(obj.ReferenceList)
}

func TestClientSetGetObjectAsync(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	group, ctx := errgroup.WithContext(context.Background())

	const jobs = 4096

	ch := make(chan datastor.Object, jobs)

	for i := 0; i < jobs; i++ {
		i := i
		group.Go(func() error {
			key := []byte(fmt.Sprintf("key#%d", i+1))
			data := make([]byte, mathRand.Int31n(4096)+1)
			rand.Read(data)

			refList := make([]string, mathRand.Int31n(16)+1)
			for i := range refList {
				id := make([]byte, mathRand.Int31n(128)+1)
				rand.Read(id)
				refList[i] = string(id)
			}

			object := datastor.Object{
				Key:           key,
				Data:          data,
				ReferenceList: refList,
			}

			err := client.SetObject(object)
			if err != nil {
				return fmt.Errorf("set error for key %q: %v", object.Key, err)
			}

			select {
			case ch <- object:
			case <-ctx.Done():
			}

			return nil
		})

		group.Go(func() error {
			var object datastor.Object
			select {
			case object = <-ch:
			case <-ctx.Done():
				return nil
			}

			outputObject, err := client.GetObject(object.Key)
			if err != nil {
				return fmt.Errorf("get error for key %q: %v", object.Key, err)
			}

			require.NotNil(outputObject)
			require.Equal(object.Key, outputObject.Key)
			require.Len(outputObject.Data, len(object.Data))
			require.Equal(outputObject.Data, object.Data)
			require.Len(outputObject.ReferenceList, len(object.ReferenceList))
			require.Equal(outputObject.ReferenceList, object.ReferenceList)

			return nil
		})
	}

	err := group.Wait()
	require.NoError(err)
}

func TestClientSetDeleteGetObject(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	err := client.DeleteObject(nil)
	require.Equal(errNilKey, err)

	err = client.DeleteObject([]byte("foo"))
	require.NoError(err)

	err = client.SetObject(datastor.Object{Key: []byte("foo"), Data: []byte("bar")})
	require.NoError(err)

	obj, err := client.GetObject([]byte("foo"))
	require.NoError(err)
	require.NotNil(obj)
	require.Equal([]byte("foo"), obj.Key)
	require.Equal([]byte("bar"), obj.Data)
	require.Empty(obj.ReferenceList)

	err = client.DeleteObject([]byte("foo"))
	require.NoError(err)

	obj, err = client.GetObject([]byte("foo"))
	require.Equal(datastor.ErrKeyNotFound, err)
	require.Nil(obj)
}

func TestClientGetObjectStatus(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	_, err := client.GetObjectStatus(nil)
	require.Equal(errNilKey, err)

	status, err := client.GetObjectStatus([]byte("foo"))
	require.NoError(err)
	require.Equal(datastor.ObjectStatusMissing, status)

	err = client.SetObject(datastor.Object{Key: []byte("foo"), Data: []byte("bar")})
	require.NoError(err)

	status, err = client.GetObjectStatus([]byte("foo"))
	require.NoError(err)
	require.Equal(datastor.ObjectStatusOK, status)
}

func TestClientExistObject(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	_, err := client.ExistObject(nil)
	require.Equal(errNilKey, err)

	exists, err := client.ExistObject([]byte("foo"))
	require.NoError(err)
	require.False(exists)

	err = client.SetObject(datastor.Object{Key: []byte("foo"), Data: []byte("bar")})
	require.NoError(err)

	exists, err = client.ExistObject([]byte("foo"))
	require.NoError(err)
	require.True(exists)
}

func TestClientGetNamespace(t *testing.T) {
	client := NewClient()
	require.Panics(t, func() {
		client.GetNamespace()
	}, "this method is not implemented for the in-memory client")
}

func TestClientListObjectKeys(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	ch, err := client.ListObjectKeyIterator(context.Background())
	require.Error(err)
	require.Nil(ch)

	err = client.SetObject(datastor.Object{Key: []byte("foo"), Data: []byte("bar")})
	require.NoError(err)
	err = client.SetObject(datastor.Object{Key: []byte("bar"), Data: []byte("foo")})
	require.NoError(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err = client.ListObjectKeyIterator(ctx)
	require.NoError(err)
	require.NotNil(ch)

	keys := map[string]struct{}{
		"foo": struct{}{},
		"bar": struct{}{},
	}
	for result := range ch {
		require.NoError(result.Error)
		require.NotNil(result.Key)

		key := string(result.Key)
		_, ok := keys[key]
		require.True(ok)
		delete(keys, key)
	}

	require.Empty(keys)

	select {
	case _, open := <-ch:
		require.False(open)
	case <-time.After(time.Millisecond * 500):
	}
}

func TestClientSetReferenceList(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	err := client.SetReferenceList(nil, nil)
	require.Equal(errNilKey, err)

	err = client.SetReferenceList([]byte("foo"), nil)
	require.Equal(errNilRefList, err)

	err = client.SetReferenceList([]byte("foo"), []string{})
	require.Equal(errNilRefList, err)

	err = client.SetReferenceList([]byte("foo"), []string{"user1"})
	require.NoError(err)
}

func TestClientGetReferenceList(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	refList, err := client.GetReferenceList(nil)
	require.Equal(errNilKey, err)
	require.Empty(refList)

	refList, err = client.GetReferenceList([]byte("foo"))
	require.Equal(datastor.ErrKeyNotFound, err)
	require.Empty(refList)

	err = client.SetReferenceList([]byte("foo"), []string{"user1"})
	require.NoError(err)

	refList, err = client.GetReferenceList([]byte("foo"))
	require.NoError(err)
	require.Equal([]string{"user1"}, refList)
}

func TestClientGetReferenceCount(t *testing.T) {
	require := require.New(t)

	client := NewClient()

	refCount, err := client.GetReferenceCount(nil)
	require.Equal(errNilKey, err)
	require.Equal(int64(0), refCount)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(0), refCount)

	err = client.SetReferenceList([]byte("foo"), []string{"user1"})
	require.NoError(err)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(1), refCount)
}

func TestClientAppendToReferenceList(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	err := client.AppendToReferenceList(nil, nil)
	require.Equal(errNilKey, err)
	err = client.AppendToReferenceList([]byte("foo"), nil)
	require.Equal(errNilRefList, err)

	refCount, err := client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(0), refCount)

	err = client.AppendToReferenceList([]byte("foo"), []string{"foo"})
	require.NoError(err)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(1), refCount)

	err = client.AppendToReferenceList([]byte("foo"), []string{"foo", "bar"})
	require.NoError(err)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(3), refCount)

	refList, err := client.GetReferenceList([]byte("foo"))
	require.NoError(err)
	require.Equal([]string{"foo", "foo", "bar"}, refList)
}

func TestClientDeleteFromReferenceList(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	refCount, err := client.DeleteFromReferenceList(nil, nil)
	require.Equal(errNilKey, err)
	require.Equal(int64(0), refCount)
	refCount, err = client.DeleteFromReferenceList([]byte("foo"), nil)
	require.Equal(errNilRefList, err)
	require.Equal(int64(0), refCount)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(0), refCount)

	refCount, err = client.DeleteFromReferenceList([]byte("foo"), []string{"foo"})
	require.NoError(err)
	require.Equal(int64(0), refCount)

	err = client.SetReferenceList([]byte("foo"), []string{"foo", "baz"})
	require.NoError(err)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(2), refCount)

	refCount, err = client.DeleteFromReferenceList([]byte("foo"), []string{"foo", "bar"})
	require.NoError(err)
	require.Equal(int64(1), refCount)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(1), refCount)

	refCount, err = client.DeleteFromReferenceList([]byte("foo"), []string{"baz", "que"})
	require.NoError(err)
	require.Equal(int64(0), refCount)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(0), refCount)

	refList, err := client.GetReferenceList([]byte("foo"))
	require.Equal(datastor.ErrKeyNotFound, err)
	require.Empty(refList)
}

func TestClientDeleteReferenceList(t *testing.T) {
	require := require.New(t)

	client := NewClient()
	defer func() {
		err := client.Close()
		require.NoError(err)
	}()

	err := client.DeleteReferenceList(nil)
	require.Equal(errNilKey, err)

	err = client.DeleteReferenceList([]byte("foo"))
	require.NoError(err)

	err = client.SetReferenceList([]byte("foo"), []string{"foo", "baz"})
	require.NoError(err)

	refCount, err := client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(2), refCount)

	err = client.DeleteReferenceList([]byte("foo"))
	require.NoError(err)

	refCount, err = client.GetReferenceCount([]byte("foo"))
	require.NoError(err)
	require.Equal(int64(0), refCount)
}
