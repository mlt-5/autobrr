// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build tools

package mobile

// gomobile bind needs golang.org/x/mobile/bind in the module it binds.
import _ "golang.org/x/mobile/bind"
