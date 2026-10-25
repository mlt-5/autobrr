// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

/// ExportedFile is a file written for the share sheet.
struct ExportedFile: Identifiable {
    let url: URL

    var id: URL { url }
}

/// ShareSheet presents the system share sheet, so a file can be saved to Files or sent to another app.
struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]

    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }

    func updateUIViewController(_ controller: UIActivityViewController, context: Context) {}
}
