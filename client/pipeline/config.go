package pipeline

import (
	"errors"
	"runtime"

	"github.com/zero-os/0-stor/client/datastor"

	"github.com/zero-os/0-stor/client/components/crypto"
	"github.com/zero-os/0-stor/client/components/storage"
	"github.com/zero-os/0-stor/client/pipeline/processing"
)

var (
	// DefaultJobCount is the JobCount,
	// while creating a config-based pipeline, in case a jobCount lower than 1 is given.
	DefaultJobCount = runtime.NumCPU() * 2
)

// TODO:
//   + Remove Close from storage.ObjectStorage and remove Close from cluster interface
//   + define new ClusterCloser interface, which will only be used for client code
//   + revive old distribution component code
//   + make all unit tests pass again
//   + add pipeline unit tests (test all different production-like pipeline combinations),
//     using a config-based creation
//   + improve error code
//   + push pipeline code as it is, non-implemented
//   + refactor all packages and integrate pipeline code into client + client-config, such that we have:
//          + client/pipeline <- { /processing, /storage, /crypto }
//          + remove entire components package
//          + integrate new client config into zstor (command-line) client
//          + ensure all unit tests pass
//  + finish PR (has 2 commits)
//
// ... once this is all done, we can improve the API of the 0-stor main client,
// and after that the hard work is done. than it's just some polishing,
// and add missing unit tests where needed.
func NewPipeline(cfg Config, cluster datastor.Cluster, jobCount int) (Pipeline, error) {
	if cluster == nil {
		panic("no datastor cluster given")
	}
	// create processor constructor
	var (
		err error
		pc  ProcessorConstructor
	)
	if cfg.Compression.Mode != processing.CompressionModeDisabled {
		pc = func() (processing.Processor, error) {
			return processing.NewCompressorDecompressor(
				cfg.Compression.Type,
				cfg.Compression.Mode)
		}
	}
	if len(cfg.Encryption.PrivateKey) != 0 {
		if pc == nil {
			// only use encryption
			pc = func() (processing.Processor, error) {
				return processing.NewEncrypterDecrypter(
					cfg.Encryption.Type,
					cfg.Encryption.PrivateKey)
			}
		} else {
			// use compression and encryption
			pc = func() (processing.Processor, error) {
				cd, err := pc()
				if err != nil {
					return nil, err
				}
				ed, err := processing.NewEncrypterDecrypter(
					cfg.Encryption.Type,
					cfg.Encryption.PrivateKey)
				if err != nil {
					return nil, err
				}
				return processing.NewProcessorChain([]processing.Processor{cd, ed}), nil
			}
		}
	}
	if pc == nil {
		// no processor used, default to the NopProcessor
		pc = func() (processing.Processor, error) { return processing.NopProcessor{}, nil }
	}
	// test our processor constructor, so we know for sure it works
	_, err = pc()
	if err != nil {
		return nil, err
	}

	// default job count if needed
	if jobCount <= 0 {
		jobCount = DefaultJobCount
	}

	// create storage
	var os storage.ObjectStorage
	if cfg.Distribution.DataShardCount <= 0 {
		if cfg.Distribution.ParityShardCount >= 0 {
			return nil, errors.New("illegal distribution: parity shard count defined, " +
				"while data shard count is undefined")
		}

		// use a non-distributed, random storage
		os, err = storage.NewRandomObjectStorage(cluster)
		if err != nil {
			return nil, err
		}
	} else {
		if cfg.Distribution.ParityShardCount <= 0 {
			// use a replication storage
			os, err = storage.NewReplicatedObjectStorage(
				cluster, cfg.Distribution.DataShardCount, jobCount)
			if err != nil {
				return nil, err
			}
		} else {
			// use a distribution (erasure code) storage
			os, err = storage.NewDistributedObjectStorage(
				cluster,
				cfg.Distribution.DataShardCount, cfg.Distribution.ParityShardCount,
				jobCount)
			if err != nil {
				return nil, err
			}
		}
	}

	// create the hasher constructor
	hashPrivateKey := cfg.Hashing.PrivateKey
	if len(hashPrivateKey) == 0 {
		hashPrivateKey = cfg.Encryption.PrivateKey
	}
	hc := func() (crypto.Hasher, error) {
		return crypto.NewHasher(
			cfg.Hashing.Type,
			hashPrivateKey,
		)
	}
	// test our counter, so we know for sure it works
	_, err = hc()
	if err != nil {
		return nil, err
	}

	// return a sequential pipeline
	if cfg.ChunkSize <= 0 {
		return NewSingleObjectPipeline(hc, pc, os), nil
	}

	// return a async splitter pipeline
	return NewAsyncSplitterPipeline(hc, pc, os, cfg.ChunkSize, jobCount), nil
}

// Config is used to configure and create a pipeline.
// While a pipeline can be manually created,
// or even defined using a custom implementation,
// creating a pipeline using this Config is the easiest
// and most recommended way to create a config.
//
// When creating a pipeline as part of a 0-stor client,
// this config will be integrated as part of that client's config,
// this is for example the case when using client in the root client package,
// which is also the client used by the zstor command-line client/tool.
//
// Note that you are required to use the same config at all times,
// for the same data (blocks). Data that was written with one config,
// are not guaranteed to be readable when using another config,
// and in most likelihood it is not possible at all.
//
// With that said, make sure to keep your config stored securely,
// as your data might not be recoverable if you lose this.
// This is definitely the case in case you lose any credentials,
// such as a private key used for encryption (and hashing).
type Config struct {
	// ChunkSize defines the size of chunks,
	// all the to be written objects should be split into.
	// If the ChunkSize has a value of 0 or lower, no object will be split
	// into multiple objects prior to writing.
	ChunkSize int

	// Hashing can not be disabled, as it is an essential part of the pipeline.
	// The keys of all stored blocks (in zstordb), are generated and
	// are equal to the checksum/signature of that block's (binary) data.
	//
	// While you cannot disable the hashing, you can however configure it.
	// Both the type of hashing algorithm can be chosen,
	// as well as the private key used to do the crypto-hashing.
	//
	// When no private key is available, this algorithm will generate a (crypto) checksum.
	// If, as recommended, a private key is available, a signature will be produced instead.
	Hashing struct {
		// The type of (crypto) hashing algorithm to use.
		// The string value (representing the hashing algorithm type), is case-insensitive.
		//
		// By default SHA_256 is used.
		// All standard types available are: SHA_256, SHA_512, Blake2b_256, Blake2b_512
		//
		// In case you've registered a custom hashing algorithm,
		// or have overridden a standard hashing algorithm, using `crypto.RegisterHasher`
		// you'll be able to use that registered hasher, by providing its (stringified) type here.
		Type crypto.HashType `json:"type" yaml:"type"`

		// PrivateKey is used to authorize the hash, proving ownership.
		// If not given, and you do use Encryption for all your data blocks,
		// as is recommended, the private key configured for Encryption
		// will also be used for the hashing (generation of block keys).
		//
		// It is recommend to have a private key available for hashing,
		// as this will make your hashing more secure and decrease the
		// chance of tamparing by a third party.
		//
		// Whether this private key is explicitly configured here,
		// or it is a shared key, and borrowed from the Encryption configuration,
		// is not as important, as your data will anyhow be
		// visible to an attacker as soon as it gained access to the Encryption's private key,
		// no matter if the Hashing private key is different or not.
		//
		// Hence this property should only really be used in case if for some reason,
		// you need/want a different private key for both hashing and encryption,
		// or in case for an even weirder reason, you want crypto-hashing,
		// while disabling encryption for the storage of the data (blocks).
		PrivateKey []byte `json:"private_key" yaml:"private_key"`
	}

	// Compressor Processor Configuration,
	// disabled by default.
	Compression struct {
		// Mode defines the compression mode to use.
		// Note that not all compression algorithms might support all modes,
		// in which case they will fall back to the closest mode that is supported.
		// If this happens a warning will be logged.
		//
		// When no mode is defined (or an explicit empty string is defined),
		// no compressor will be created.
		//
		// All standard compression modes available are: default, best_speed, best_compression
		Mode processing.CompressionMode `json:"mode" yaml:"mode"`

		// The type of compression algorithm to use,
		// defining both the compressing and decompressing logic.
		// The string value (representing the compression algorithm type), is case-insensitive.
		//
		// The default compression type is: Snappy
		// All standard compression types available are: Snappy, LZ4, GZip
		//
		// In case you've registered a custom compression algorithm,
		// or have overridden a standard compression algorithm, using `processing.RegisterCompressorDecompressor`
		// you'll be able to use that compressor-decompressor, by providing its (stringified) type here.
		Type processing.CompressionType `json:"type" yaml:"type"`
	} `json:"compression" yaml:"compression"`

	// Encryption algorithm configuration, defining, when enabled,
	// how to encrypt all blocks prior to writing, and decrypt them once again when reading.
	// When both Compression and Encryption is configured and used,
	// the compressed blocks will be encrypting when writing,
	// and the decrypted blocks will be decompressed.
	//
	// Encryption is disabled by default and can be enabled by providing
	// a valid private key. Optionally you can also define a different encryption algorithm on top of that.
	//
	// It is recommended to use encryption, and do so using the AES_256 algorithm.
	Encryption struct {
		// Private key, the specific required length
		// is defined by the type of Encryption used.
		//
		// This key will also used by the crypto-hashing algorithm given,
		// if you did not define a separate key within the hashing configuration.
		PrivateKey []byte `json:"private_key" yaml:"private_key"`

		// The type of encryption algorithm to use,
		// defining both the encrypting and decrypting logic.
		// The string value (representing the encryption algorithm type), is case-insensitive.
		//
		// By default no type is used, disabling encryption,
		// encryption gets enabled as soon as a private key gets defined.
		// All standard types available are: AES
		//
		// Valid Key sizes for AES are: 16, 24 and 32 bytes
		// The recommended private key size is 32 bytes, this will select/use AES_256.
		//
		// In case you've registered a custom encryption algorithm,
		// or have overridden a standard encryption algorithm, using `processing.RegisterEncrypterDecrypter`
		// you'll be able to use that encrypter-decrypting, by providing its (stringified) type here.
		Type processing.EncryptionType `json:"type" yaml:"yaml"`
	} `json:"encryption" yaml:"encryption"`

	// Shards defines the list of zstordb servers, used as a single cluster.
	//
	// At least Distribution.DataShardCount+Distribution.ParityShardCount is required
	Shards []string `json:"shards" yaml:"shards"`

	// Distribution defines how all blocks should-be/are distributed.
	// These properties are optional, and when not given,
	// it will simply store each block on a single shard (zstordb server),
	// by default. Thus if you do not specify any of these properties,
	// (part of) your data is lost, as soon as
	// a shard used to store (part of) becomes unavailable.
	//
	// If only DataShardCount is given AND positive,
	// for all blocks, each block will be stored onto multiple shards (replication).
	//
	// If both DataShardCount is positive and ParityShardCount is positive as well,
	// erasure-code distribution is used, reducing the performance of the zstor client,
	// but increasing the data redendency (resilience) of the stored blocks.
	//
	// All other possible property combinations are illegal
	// and will result in an invalid Pipeline configuration,
	// returning an error upon creation.
	Distribution struct {
		// Number of data shards to use for each stored block.
		// If only the DataShardCount is given, replication is used.
		// If used in combination with a positive ParityCount,
		// erasure-code distribution is used.
		DataShardCount int `json:"data_shards" yaml:"data_shards"`

		// Number of parity shards to use for each stored block.
		// When both this value and DataShardCount are positive, each,
		// erasure-code distribution is used to read and write all blocks.
		// When ParityCount is positive and DataShardCount is zero or lower,
		// it will invalidate this Config, as this is not an acceptable combination.
		//
		// Of all available data shards (defined by DataShardCount),
		// you can lose up to ParityCount of shards.
		// Meaning that if you have a DatashardCount of 10, and a ParityShardCount of 3,
		// you're data is read-able and repair-able as long
		// as you have 7 data shards, or more, available. In this example,
		// as soon as you only 6 data shards or less available, and have the others one unavailable,
		// your data will no longer be available and you will suffer from (partial) data loss.
		//
		// Note that when using this configuration,
		// make sure that you have at least DataShardCount+ParityShardCount shards available,
		// meaning that in our example of above you would need at least 13 shards.
		// However, it would be even safer if you could have make shards available than the minimum,
		// a this would mean you can still write and repair in case you lose some shards.
		// If, in our example, you would have less than 13 shards, but more than 6,
		// you would still be able to read data, writing and repairing would no longer be possible.
		// There in our example it would be better if we provide more than 13 shards.
		ParityShardCount int `json:"parity_count" yaml:"parity_count"`
	} `json:"distribution" yaml:"distribution"`
}
