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

package openvpn

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorIsReturnedOnBadBinaryPath(t *testing.T) {
	assert.Error(t, CheckOpenvpnBinary("non-existent-binary"))
}

func TestErrorIsReturnedOnExitCodeZero(t *testing.T) {
	assert.Error(t, CheckOpenvpnBinary("testdata/exit-with-zero.sh"))
}

func TestNoErrorIsReturnedOnExitCodeOne(t *testing.T) {
	assert.NoError(t, CheckOpenvpnBinary("testdata/exit-with-one.sh"))
}
