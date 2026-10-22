// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

/// TextRow edits a text field. Multiline rows grow for comma or newline separated lists.
struct TextRow: View {
    let label: String
    @Binding var text: String
    var prompt = ""
    var multiline = false

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.caption)
                .foregroundStyle(.secondary)

            TextField(prompt, text: $text, axis: multiline ? .vertical : .horizontal)
                .lineLimit(multiline ? 1...6 : 1...1)
                .autocorrectionDisabled()
                .textInputAutocapitalization(.never)
        }
    }
}

struct NumberRow: View {
    let label: String
    @Binding var value: Int

    var body: some View {
        LabeledContent(label) {
            TextField("0", value: $value, format: .number)
                .keyboardType(.numberPad)
                .multilineTextAlignment(.trailing)
                .frame(maxWidth: 120)
        }
    }
}

/// MultiSelectRow picks any number of values from a fixed list on its own screen.
struct MultiSelectRow: View {
    let label: String
    let options: [String]
    @Binding var selection: [String]
    var titles: [String: String] = [:]

    var body: some View {
        NavigationLink {
            List(options, id: \.self) { option in
                Button {
                    if let index = selection.firstIndex(of: option) {
                        selection.remove(at: index)
                    } else {
                        selection.append(option)
                    }
                } label: {
                    HStack {
                        Text(titles[option] ?? option)
                            .foregroundStyle(Color.primary)
                        Spacer()
                        if selection.contains(option) {
                            Image(systemName: "checkmark")
                                .foregroundStyle(Color.accentColor)
                        }
                    }
                }
            }
            .navigationTitle(label)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                if !selection.isEmpty {
                    Button("Clear") { selection = [] }
                }
            }
        } label: {
            LabeledContent(label) {
                Text(selection.isEmpty ? "Any" : selection.map { titles[$0] ?? $0 }.joined(separator: ", "))
                    .lineLimit(1)
            }
        }
    }
}
