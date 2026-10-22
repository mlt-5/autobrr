// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import SwiftUI

/// SettingsView lists the settings sections. Sections not yet rebuilt natively open the web UI at their route.
struct SettingsView: View {
    let engine: MobileEngine
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
                    web("Indexers", "/settings/indexers", "antenna.radiowaves.left.and.right")
                    web("IRC", "/settings/irc", "bubble.left.and.bubble.right")
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
                    web("Releases", "/settings/releases", "tray.full")
                }
            }
            .navigationTitle("Settings")
        }
    }

    private func web(_ title: String, _ path: String, _ symbol: String) -> some View {
        NavigationLink {
            WebView(handler: SchemeHandler(engine: engine), path: path)
                .navigationTitle(title)
                .navigationBarTitleDisplayMode(.inline)
        } label: {
            Label(title, systemImage: symbol)
        }
    }
}
