// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct SettingsView: View {
    let client: EngineClient

    var body: some View {
        NavigationStack {
            List {
                Section {
                    NavigationLink {
                        ApplicationSettingsView(client: client)
                    } label: {
                        Label("Application", systemImage: "app.badge")
                    }
                }

                Section {
                    NavigationLink {
                        IndexersView(client: client)
                    } label: {
                        Label("Indexers", systemImage: "antenna.radiowaves.left.and.right")
                    }
                    NavigationLink {
                        IrcSettingsView(client: client)
                    } label: {
                        Label("IRC", systemImage: "bubble.left.and.bubble.right")
                    }
                    NavigationLink {
                        FeedsView(client: client)
                    } label: {
                        Label("Feeds", systemImage: "dot.radiowaves.up.forward")
                    }
                    NavigationLink {
                        ListsView(client: client)
                    } label: {
                        Label("Lists", systemImage: "list.bullet.rectangle")
                    }
                    NavigationLink {
                        NotificationsView(client: client)
                    } label: {
                        Label("Notifications", systemImage: "bell")
                    }
                    NavigationLink {
                        ProxiesView(client: client)
                    } label: {
                        Label("Proxies", systemImage: "network")
                    }
                    NavigationLink {
                        ReleaseSettingsView(client: client)
                    } label: {
                        Label("Releases", systemImage: "tray.full")
                    }
                }
            }
            .navigationTitle("Settings")
        }
    }
}
