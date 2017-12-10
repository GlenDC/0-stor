package client

import (
	"fmt"
	"sync"

	"github.com/zero-os/0-stor/client/metastor"
)

// SetReferenceList replace the complete reference list for the object pointed by key
func (c *Client) SetReferenceList(key []byte, refList []string) error {
	md, err := c.metaCli.GetMetadata(key)
	if err != nil {
		return err
	}
	return c.SetReferenceListWithMeta(md, refList)
}

// SetReferenceListWithMeta is the same as SetReferenceList but take metadata instead of key
// as argument
func (c *Client) SetReferenceListWithMeta(md *metastor.Data, refList []string) error {
	return c.updateRefListWithMeta(md, refList, refListOpSet)
}

// AppendReferenceList adds some reference to the reference list of the object pointed by key
func (c *Client) AppendReferenceList(key []byte, refList []string) error {
	md, err := c.metaCli.GetMetadata(key)
	if err != nil {
		return err
	}
	return c.AppendReferenceListWithMeta(md, refList)
}

// AppendReferenceListWithMeta is the same as AppendReferenceList but take metadata instead of key
// as argument
func (c *Client) AppendReferenceListWithMeta(md *metastor.Data, refList []string) error {
	return c.updateRefListWithMeta(md, refList, refListOpAppend)
}

// RemoveReferenceList removes some reference from the reference list of the object pointed by key.
// It wont return error in case of the object doesn't have some elements of the `refList`.
func (c *Client) RemoveReferenceList(key []byte, refList []string) error {
	md, err := c.metaCli.GetMetadata(key)
	if err != nil {
		return err
	}
	return c.RemoveReferenceListWithMeta(md, refList)
}

// RemoveReferenceListWithMeta is the same as RemoveReferenceList but take metadata
// instead of key as argument
func (c *Client) RemoveReferenceListWithMeta(md *metastor.Data, refList []string) error {
	return c.updateRefListWithMeta(md, refList, refListOpRemove)
}

func (c *Client) updateRefListWithMeta(md *metastor.Data, refList []string, op int) error {
	for _, chunk := range md.Chunks {

		var (
			wg    sync.WaitGroup
			errCh = make(chan error, len(chunk.Shards))
		)

		wg.Add(len(chunk.Shards))
		for _, shard := range chunk.Shards {
			go func(shard string) {
				defer wg.Done()

				// get stor client
				storCli, err := c.getStor(shard)
				if err != nil {
					errCh <- err
					return
				}

				// do the work
				switch op {
				case refListOpSet:
					err = storCli.SetReferenceList(chunk.Key, refList)
				case refListOpAppend:
					err = storCli.AppendToReferenceList(chunk.Key, refList)
				case refListOpRemove:
					// TODO: return the count value to the user?!
					_, err = storCli.DeleteFromReferenceList(chunk.Key, refList)
				default:
					err = fmt.Errorf("wrong operation: %v", op)
				}
				if err != nil {
					errCh <- err
					return
				}
			}(shard)
		}

		wg.Wait()

		if len(errCh) > 0 {
			err := <-errCh
			return err
		}
	}
	return nil
}

const (
	_ = iota
	refListOpSet
	refListOpAppend
	refListOpRemove
)
