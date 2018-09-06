/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
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

package discovery

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/identity/registry"
	"github.com/mysteriumnetwork/node/openvpn/discovery"
	"github.com/mysteriumnetwork/node/server"
	dto_discovery "github.com/mysteriumnetwork/node/service_discovery/dto"
)

// Discovery structure holds discovery service state
type Discovery struct {
	identityRegistry            registry.IdentityRegistry
	ownIdentity                 common.Address
	registrationDataProvider    registry.RegistrationDataProvider
	mysteriumClient             server.Client
	signer                      identity.Signer
	proposal                    dto_discovery.ServiceProposal
	proposalStatusChan          chan ProposalStatus
	status                      ProposalStatus
	proposalAnnouncementStopped *sync.WaitGroup
	unsubscribe                 func()
	stop                        func()

	sync.RWMutex
}

// NewService creates new discovery service
func NewService(
	identityRegistry registry.IdentityRegistry,
	ownIdentity identity.Identity,
	provider registry.RegistrationDataProvider,
	mysteriumClient server.Client,
	createSigner identity.SignerFactory,
) *Discovery {
	signer := createSigner(identity.FromAddress(ownIdentity.Address))

	return &Discovery{
		identityRegistry,
		common.HexToAddress(ownIdentity.Address),
		provider,
		mysteriumClient,
		signer,
		dto_discovery.ServiceProposal{},
		make(chan ProposalStatus),
		StatusUndefined,
		&sync.WaitGroup{},
		func() {},
		func() {},
		sync.RWMutex{},
	}
}

// GenertateServiceProposalWithLocation service proposal wrapper method for openvpn NewServiceProposalWithLocation method
func (d *Discovery) GenertateServiceProposalWithLocation(
	providerID identity.Identity,
	providerContact dto_discovery.Contact,
	serviceLocation dto_discovery.Location,
	protocol string,
) dto_discovery.ServiceProposal {
	d.proposal = discovery.NewServiceProposalWithLocation(providerID, providerContact, serviceLocation, protocol)
	return d.proposal
}
