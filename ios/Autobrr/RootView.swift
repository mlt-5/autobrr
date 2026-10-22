// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import SwiftUI

/// RootView holds the app's tabs.
struct RootView: View {
    private let engine: MobileEngine
    private let client: EngineClient

    init(engine: MobileEngine) {
        self.engine = engine
        self.client = EngineClient(engine: engine)
    }

    var body: some View {
        TabView {
            DashboardView(client: client)
                .tabItem { Label("Dashboard", systemImage: "chart.bar") }

            FiltersView(client: client)
                .tabItem { Label("Filters", systemImage: "line.3.horizontal.decrease.circle") }

            ReleasesView(client: client)
                .tabItem { Label("Releases", systemImage: "tray.full") }

            LogsView(client: client)
                .tabItem { Label("Logs", systemImage: "text.alignleft") }

            SettingsView(engine: engine, client: client)
                .tabItem { Label("Settings", systemImage: "gearshape") }
        }
    }
}
