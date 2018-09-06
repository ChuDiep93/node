/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package registry

import (
	"context"
	"time"

	log "github.com/cihub/seelog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mysteriumnetwork/payments/registry"
	"github.com/mysteriumnetwork/payments/registry/generated"
)

const logPrefix = "[registry] "

// Register type combines different IdentityRegistry interfaces under single type
type Register struct {
	callerSession *generated.IdentityRegistryCallerSession
	filterer      *generated.IdentityRegistryFilterer
}

// IdentityRegistry enables identity registration actions
type IdentityRegistry interface {
	IsRegistered(identity common.Address) (bool, error)
	SubscribeToRegistrationEvent(providerAddress common.Address) (registeredEvent chan RegistrationEvent, unsubscribe func())
}

// RegistrationDataProvider provides registration information for given identity required to register it on blockchain
type RegistrationDataProvider interface {
	ProvideRegistrationData(identity common.Address) (*registry.RegistrationData, error)
}

type keystoreRegistrationDataProvider struct {
	ks *keystore.KeyStore
}

// ProvideRegistrationData returns registration data needed to register given identity with payments contract
func (kpg *keystoreRegistrationDataProvider) ProvideRegistrationData(identity common.Address) (*registry.RegistrationData, error) {
	identityHolder := registry.FromKeystore(kpg.ks, identity)

	return registry.CreateRegistrationData(identityHolder)
}

// NewRegistrationDataProvider creates registration data provider backed up by identity which is managed by keystore
func NewRegistrationDataProvider(ks *keystore.KeyStore) RegistrationDataProvider {
	return &keystoreRegistrationDataProvider{
		ks: ks,
	}
}

// NewIdentityRegistry creates identity registry service which uses blockchain for information
func NewIdentityRegistry(contractBackend bind.ContractBackend, registryAddress common.Address) (IdentityRegistry, error) {
	contract, err := generated.NewIdentityRegistryCaller(registryAddress, contractBackend)
	if err != nil {
		return nil, err
	}

	filterer, err := generated.NewIdentityRegistryFilterer(registryAddress, contractBackend)
	if err != nil {
		return nil, err
	}

	return &Register{
		&generated.IdentityRegistryCallerSession{
			Contract: contract,
			CallOpts: bind.CallOpts{
				Pending: false, //we want to find out true registration status - not pending transactions
			},
		},
		filterer,
	}, nil
}

// IsRegistered returns identity registration status within payments contract
func (register *Register) IsRegistered(identity common.Address) (bool, error) {
	return register.callerSession.IsRegistered(identity)
}

// RegistrationEvent describes registration events
type RegistrationEvent int

// Possible registration events
const (
	Registered RegistrationEvent = 0
	Cancelled  RegistrationEvent = 1
)

// SubscribeToRegistrationEvent returns registration event if given providerAddress was registered within payments contract
func (register *Register) SubscribeToRegistrationEvent(providerAddress common.Address) (registrationEvent chan RegistrationEvent, unsubscribe func()) {
	registrationEvent = make(chan RegistrationEvent)

	stopLoop := make(chan bool)
	unsubscribe = func() {
		// cancel (stop) identity registration loop
		stopLoop <- true
	}

	identities := []common.Address{providerAddress}

	filterOps := &bind.FilterOpts{
		Start:   0,
		End:     nil,
		Context: context.Background(),
	}

	go func() {
		for {
			select {
			case <-stopLoop:
				registrationEvent <- Cancelled
				return
			case <-time.After(1 * time.Second):
				logIterator, err := register.filterer.FilterRegistered(filterOps, identities)
				if err != nil {
					registrationEvent <- Cancelled
					log.Error(err)
					return
				}
				if logIterator == nil {
					registrationEvent <- Cancelled
					return
				}
				for {
					next := logIterator.Next()
					if next {
						registrationEvent <- Registered
					} else {
						err = logIterator.Error()
						if err != nil {
							log.Error(err)
						}
						break
					}
				}
				log.Trace(logPrefix, "no identity registration, sleeping for 1s")
			}
		}
	}()
	return
}
