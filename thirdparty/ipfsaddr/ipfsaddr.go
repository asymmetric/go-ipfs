package ipfsaddr

import (
	"errors"

	ma "gx/ipfs/QmcyqRMCAXVtYPS4DiBrA7sezL9rRGfW8Ctx7cywL4TXJj/go-multiaddr"

	path "github.com/ipfs/go-ipfs/path"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmdS9KpbDyPrieswibZhkod1oXqRwZJrUPzxCofAMWpFGq/go-libp2p-peer"
)

var log = logging.Logger("ipfsaddr")

// ErrInvalidAddr signals an address is not a valid IPFS address.
var ErrInvalidAddr = errors.New("invalid IPFS address")

type IPFSAddr interface {
	ID() peer.ID
	Multiaddr() ma.Multiaddr
	Transport() ma.Multiaddr
	String() string
	Equal(b interface{}) bool
}

type ipfsAddr struct {
	ma ma.Multiaddr
	id peer.ID
}

func (a ipfsAddr) ID() peer.ID {
	return a.id
}

func (a ipfsAddr) Multiaddr() ma.Multiaddr {
	return a.ma
}

func (a ipfsAddr) Transport() ma.Multiaddr {
	return Transport(a)
}

func (a ipfsAddr) String() string {
	return a.ma.String()
}

func (a ipfsAddr) Equal(b interface{}) bool {
	if ib, ok := b.(IPFSAddr); ok {
		return a.Multiaddr().Equal(ib.Multiaddr())
	}
	if mb, ok := b.(ma.Multiaddr); ok {
		return a.Multiaddr().Equal(mb)
	}
	return false
}

// ParseString parses a string representation of an address into an IPFSAddr
func ParseString(str string) (a IPFSAddr, err error) {
	if str == "" {
		return nil, ErrInvalidAddr
	}

	m, err := ma.NewMultiaddr(str)
	if err != nil {
		return nil, err
	}

	return ParseMultiaddr(m)
}

// ParseMultiaddr parses a multiaddr into an IPFSAddr
func ParseMultiaddr(m ma.Multiaddr) (a IPFSAddr, err error) {
	// never panic.
	defer func() {
		if r := recover(); r != nil {
			log.Debug("recovered from panic: ", r)
			a = nil
			err = ErrInvalidAddr
		}
	}()

	if m == nil {
		return nil, ErrInvalidAddr
	}

	// make sure it's an IPFS addr
	parts := ma.Split(m)
	if len(parts) < 1 {
		return nil, ErrInvalidAddr
	}
	ipfspart := parts[len(parts)-1] // last part
	if ipfspart.Protocols()[0].Code != ma.P_IPFS {
		return nil, ErrInvalidAddr
	}

	// make sure 'ipfs id' parses as a peer.ID
	peerIdParts := path.SplitList(ipfspart.String())
	peerIdStr := peerIdParts[len(peerIdParts)-1]
	id, err := peer.IDB58Decode(peerIdStr)
	if err != nil {
		return nil, err
	}

	return ipfsAddr{ma: m, id: id}, nil
}

func Transport(iaddr IPFSAddr) (maddr ma.Multiaddr) {
	maddr = iaddr.Multiaddr()
	split := ma.Split(maddr)
	maddr = ma.Join(split[:len(split)-1]...)
	return
}
