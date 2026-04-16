package geolocation

import (
	"fmt"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/logger"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type MaxMindClient struct {
	cfg        *config.Config
	mmdbReader *geoip2.Reader
}

func NewMaxMindClient(cfg *config.Config) (*MaxMindClient, error) {
	if module.IsEmptyString(cfg.MAX_MIND_DATABASE) {
		msg := "MAX_MIND_DATABASE variable is not set"
		fmt.Println(msg)
		logger.Log.Warn(msg)

		return nil, nil
	}

	mmdbReader, err := geoip2.Open(cfg.MAX_MIND_DATABASE)
	if err != nil {
		msg := "could not open MaxMind DB file %s: %v"
		fmt.Println(fmt.Errorf(msg, cfg.MAX_MIND_DATABASE, err))
		logger.Log.Errorf(msg, cfg.MAX_MIND_DATABASE, err)

		return nil, ErrGeoLocationClientNotInit
	}

	return &MaxMindClient{
		cfg:        cfg,
		mmdbReader: mmdbReader,
	}, nil
}

// lookupCountryCode uses the MaxMind reader to find the country code for an IP
func (client *MaxMindClient) LookupCountryCode(ipAddressStr string) (string, error) {
	if client.mmdbReader == nil || module.IsEmptyString(client.cfg.MAX_MIND_DATABASE) || module.IsEmptyString(ipAddressStr) {
		return "", nil
	}

	ipAddress := net.ParseIP(ipAddressStr)
	if ipAddress == nil {
		logger.Log.Errorf("invalid IP Address %s", ipAddressStr)

		return "", ErrInvalidIPAddress
	}

	record, err := client.mmdbReader.Country(ipAddress)
	if err != nil {
		logger.Log.Errorf("maxmind lookup error for IP %s: %v", ipAddressStr, err)

		return "", ErrLookupIPAddress
	}

	// ISO 3166-1 alpha-2 code (e.g., "US")
	return record.Country.IsoCode, nil
}
