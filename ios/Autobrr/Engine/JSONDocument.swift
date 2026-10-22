// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

/// JSONDocument is an engine object edited field by field through bindings, keeping fields no screen shows.
@MainActor
protocol JSONDocument: AnyObject {
    var values: [String: JSONValue] { get set }
}

extension JSONDocument {
    func set(_ key: String, _ value: JSONValue) {
        values[key] = value
    }

    func string(_ key: String) -> Binding<String> {
        Binding { self.values[key]?.stringValue ?? "" } set: { self.set(key, .string($0)) }
    }

    func bool(_ key: String) -> Binding<Bool> {
        Binding { self.values[key]?.boolValue ?? false } set: { self.set(key, .bool($0)) }
    }

    func int(_ key: String) -> Binding<Int> {
        Binding { Int(self.values[key]?.intValue ?? 0) } set: { self.set(key, .int(Int64($0))) }
    }

    func list(_ key: String) -> Binding<[String]> {
        Binding {
            self.values[key]?.arrayValue?.compactMap(\.stringValue) ?? []
        } set: {
            self.set(key, .array($0.map(JSONValue.string)))
        }
    }

    func string(in key: String, _ field: String) -> Binding<String> {
        Binding { self.values[key]?.objectValue?[field]?.stringValue ?? "" } set: { value in
            self.updateObject(key) { $0[field] = .string(value) }
        }
    }

    func int(in key: String, _ field: String) -> Binding<Int> {
        Binding { Int(self.values[key]?.objectValue?[field]?.intValue ?? 0) } set: { value in
            self.updateObject(key) { $0[field] = .int(Int64(value)) }
        }
    }

    /// commaList edits a comma separated string field as a list.
    func commaList(_ key: String) -> Binding<[String]> {
        Binding {
            (self.values[key]?.stringValue ?? "").split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }
        } set: {
            self.set(key, .string($0.joined(separator: ",")))
        }
    }

    func items(_ key: String) -> [[String: JSONValue]] {
        values[key]?.arrayValue?.compactMap(\.objectValue) ?? []
    }

    func setItems(_ key: String, _ items: [[String: JSONValue]]) {
        set(key, .array(items.map(JSONValue.object)))
    }

    func string(_ key: String, _ index: Int, _ field: String) -> Binding<String> {
        Binding { self.field(key, index, field)?.stringValue ?? "" } set: { value in
            self.update(key, index) { $0[field] = .string(value) }
        }
    }

    func bool(_ key: String, _ index: Int, _ field: String) -> Binding<Bool> {
        Binding { self.field(key, index, field)?.boolValue ?? false } set: { value in
            self.update(key, index) { $0[field] = .bool(value) }
        }
    }

    func int(_ key: String, _ index: Int, _ field: String) -> Binding<Int> {
        Binding { Int(self.field(key, index, field)?.intValue ?? 0) } set: { value in
            self.update(key, index) { $0[field] = .int(Int64(value)) }
        }
    }

    func list(_ key: String, _ index: Int, _ field: String) -> Binding<[String]> {
        Binding {
            self.field(key, index, field)?.arrayValue?.compactMap(\.stringValue) ?? []
        } set: { value in
            self.update(key, index) { $0[field] = .array(value.map(JSONValue.string)) }
        }
    }

    private func updateObject(_ key: String, _ change: (inout [String: JSONValue]) -> Void) {
        var object = values[key]?.objectValue ?? [:]
        change(&object)
        set(key, .object(object))
    }

    private func field(_ key: String, _ index: Int, _ field: String) -> JSONValue? {
        let items = items(key)
        return items.indices.contains(index) ? items[index][field] : nil
    }

    private func update(_ key: String, _ index: Int, _ change: (inout [String: JSONValue]) -> Void) {
        var items = items(key)
        guard items.indices.contains(index) else { return }

        change(&items[index])
        setItems(key, items)
    }
}
