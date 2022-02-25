package registration

import (
	"errors"
	"github.com/LK4D4/trylock"
	"github.com/grussorusso/serverledge/internal/config"
	"github.com/hexablock/vivaldi"
)

var UnavailableClientErr = errors.New("etcd client unavailable")
var IdRegistrationErr = errors.New("etcd error: could not complete the registration")
var KeepAliveErr = errors.New(" The system can't renew your registration key")

type Registry struct {
	Area             string
	Key              string
	Client           *vivaldi.Client
	RwMtx            trylock.Mutex
	NearbyServersMap map[string]*StatusInformation
	serversMap       map[string]*StatusInformation
	etcdCh           chan bool
}

var BASEDIR = "registry"
var TTL = config.GetInt(config.REGISTRATION_TTL, 90) // lease time in Seconds

type StatusInformation struct {
	Url            string
	AvailableMemMB int64
	AvailableCPUs  float64
	DropCount      int64
	Coordinates    vivaldi.Coordinate
}