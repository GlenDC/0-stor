package test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/zero-os/0-stor/client/itsyouonline"
	pb "github.com/zero-os/0-stor/server/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// authentication key where the JWT can be found in the GRPC's context
	authGRPCKey = "authorization"

	// test organization
	organization = "testorg"

	// test namespace
	namespace = "testnamespace"

	// test label (full iyo namespacing)
	label = "testorg_0stor_testnamespace"
)

func TestListObject(t *testing.T) {
	server, iyoCl, clean := getTestGRPCServer(t, organization)
	bufList := populateDB(t, label, server.DB())

	// create client connection
	conn, err := grpc.Dial(server.Addr(), grpc.WithInsecure())
	require.NoError(t, err, "can't connect to the server")

	defer func() {
		conn.Close()
		clean()
	}()

	cl := pb.NewObjectManagerClient(conn)
	t.Run("valid object", func(t *testing.T) {

		jwt, err := iyoCl.CreateJWT(namespace, itsyouonline.Permission{
			Read: true,
		})
		require.NoError(t, err, "fail to generate jwt")

		md := metadata.Pairs(authGRPCKey, jwt)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		stream, err := cl.List(ctx, &pb.ListObjectsRequest{Label: label})
		require.NoError(t, err, "can't send list request to server")

		objNr := 0
		for i := 0; ; i++ {

			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("error while reading stream: %v", err)
			}

			objNr++
			expectedValue, ok := bufList[string(obj.Key)]
			require.True(t, ok, fmt.Sprintf("received key that was not present in db %s", obj.GetKey()))
			assert.EqualValues(t, expectedValue, obj.GetValue())
		}
		assert.Equal(t, len(bufList), objNr)
	})

	t.Run("wrong permission", func(t *testing.T) {
		jwt, err := iyoCl.CreateJWT(namespace, itsyouonline.Permission{
			Write: true,
		})
		require.NoError(t, err, "fail to generate jwt")

		md := metadata.Pairs(authGRPCKey, jwt)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		stream, err := cl.List(ctx, &pb.ListObjectsRequest{Label: label})
		require.NoError(t, err, "failed to call List")
		for {
			_, err = stream.Recv()
			if err == io.EOF {
				break
			}

			require.Error(t, err)
			statusErr, ok := status.FromError(err)
			require.True(t, ok, "error is not valid rpc status error")
			assert.Equal(t, "JWT token doesn't contains required scopes", statusErr.Message())
			break
		}
	})

	t.Run("admin right", func(t *testing.T) {
		jwt, err := iyoCl.CreateJWT(namespace, itsyouonline.Permission{
			Admin: true,
		})
		require.NoError(t, err, "fail to generate jwt")

		md := metadata.Pairs(authGRPCKey, jwt)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		stream, err := cl.List(ctx, &pb.ListObjectsRequest{Label: label})
		require.NoError(t, err, "failed to call List")
		_, err = stream.Recv()
		require.NoError(t, err)
		stream.CloseSend()
	})
}

func TestCheckObject(t *testing.T) {
	server, iyoCl, clean := getTestGRPCServer(t, organization)
	populateDB(t, label, server.DB())

	// create client connection
	conn, err := grpc.Dial(server.Addr(), grpc.WithInsecure())
	require.NoError(t, err, "can't connect to the server")

	defer func() {
		conn.Close()
		clean()
	}()

	cl := pb.NewObjectManagerClient(conn)
	jwt, err := iyoCl.CreateJWT(namespace, itsyouonline.Permission{
		Read: true,
	})
	require.NoError(t, err, "fail to generate jwt")

	tt := []struct {
		name           string
		keys           []string
		expectedStatus pb.CheckResponse_Status
	}{
		{
			name:           "valid",
			keys:           []string{"testkey1", "testkey2", "testkey3"},
			expectedStatus: pb.CheckResponse_ok,
		},
		{
			name:           "missing",
			keys:           []string{"dontexsits"},
			expectedStatus: pb.CheckResponse_missing,
		},
	}

	for _, tc := range tt {
		md := metadata.Pairs(authGRPCKey, jwt)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		stream, err := cl.Check(ctx, &pb.CheckRequest{
			Label: label,
			Ids:   tc.keys,
		})
		require.NoError(t, err, "fail to send check request")

		n := 0
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err, "error during check response streaming")

			assert.Equal(t, tc.expectedStatus, resp.GetStatus(), fmt.Sprintf("status should be %v", tc.expectedStatus))
			n++
		}

		assert.Equal(t, len(tc.keys), n)
	}
}
