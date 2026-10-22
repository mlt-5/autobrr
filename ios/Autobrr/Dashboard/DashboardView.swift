// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Charts
import SwiftUI

struct DashboardView: View {
    let client: EngineClient

    @StateObject private var model: DashboardModel
    @StateObject private var inbox: InboxModel

    init(client: EngineClient) {
        self.client = client
        _model = StateObject(wrappedValue: DashboardModel(client: client))
        _inbox = StateObject(wrappedValue: InboxModel(client: client))
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    if let error = model.error {
                        Label(error, systemImage: "exclamationmark.triangle")
                            .font(.footnote)
                            .foregroundStyle(.red)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }

                    alertBanners

                    tiles

                    Picker("Period", selection: $model.period) {
                        ForEach(DashboardModel.Period.allCases) { Text($0.label).tag($0) }
                    }
                    .pickerStyle(.segmented)

                    card("Activity") { ActivityChart(days: model.activity) }
                    card("Downloaded") { VolumeChart(days: model.volume) }
                    card("Busiest hours") { HeatmapChart(cells: model.heatmap) }
                    card("Top indexers") { TopList(entries: model.topIndexers) }
                    card("Top filters") { TopList(entries: model.topFilters) }
                }
                .padding()
            }
            .background(Color(.systemGroupedBackground))
            .navigationTitle("Dashboard")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    NavigationLink {
                        InboxView(model: inbox, client: client)
                    } label: {
                        Image(systemName: inbox.unreadCount > 0 ? "bell.badge" : "bell")
                    }
                    .accessibilityLabel(inbox.unreadCount > 0 ? "Inbox, \(inbox.unreadCount) unread" : "Inbox")
                }
            }
            .task {
                await inbox.load()
                await inbox.watch()
            }
            .refreshable { await model.load() }
            .task(id: model.period) { await model.load() }
            .overlay {
                if !model.hasLoaded { ProgressView() }
            }
        }
    }

    @ViewBuilder
    private var alertBanners: some View {
        let irc = model.alerts.filter { $0.kind == "IRC_UNHEALTHY" }
        let lists = model.alerts.filter { $0.kind == "LIST_REFRESH_ERROR" }

        if !irc.isEmpty {
            NavigationLink {
                IrcSettingsView(client: client)
            } label: {
                banner(irc.count == 1 ? "IRC network unhealthy" : "IRC networks unhealthy", irc)
            }
        }

        if !lists.isEmpty {
            NavigationLink {
                ListsView(client: client)
            } label: {
                banner(lists.count == 1 ? "List refresh failed" : "List refreshes failed", lists)
            }
        }
    }

    private func banner(_ title: String, _ alerts: [EngineAlert]) -> some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "exclamationmark.triangle.fill")
            VStack(alignment: .leading, spacing: 2) {
                Text(title).fontWeight(.semibold)
                ForEach(alerts) { alert in
                    Text([alert.subject, alert.message ?? ""].filter { !$0.isEmpty }.joined(separator: ": "))
                        .font(.caption)
                        .lineLimit(2)
                }
            }
            Spacer()
            Image(systemName: "chevron.right")
                .font(.caption)
        }
        .foregroundStyle(.white)
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.red, in: RoundedRectangle(cornerRadius: 12))
    }

    private var tiles: some View {
        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
            tile("Filtered", model.stats?.filteredCount, .blue)
            tile("Approved", model.stats?.pushApprovedCount, .green)
            tile("Rejected", model.stats?.pushRejectedCount, .orange)
            tile("Errors", model.stats?.pushErrorCount, .red)
        }
    }

    private func tile(_ title: String, _ value: Int64?, _ color: Color) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(title)
                .font(.caption)
                .foregroundStyle(.secondary)
            Text(value.map { $0.formatted() } ?? "–")
                .font(.title2.weight(.semibold).monospacedDigit())
                .foregroundStyle(color)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(Color(.secondarySystemGroupedBackground), in: RoundedRectangle(cornerRadius: 12))
    }

    private func card<Content: View>(_ title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(title).font(.headline)
            content()
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(Color(.secondarySystemGroupedBackground), in: RoundedRectangle(cornerRadius: 12))
    }
}

private struct ActivityChart: View {
    let days: [ActivityDay]

    var body: some View {
        if days.allSatisfy({ $0.matchedCount == 0 }) {
            EmptyChart()
        } else {
            Chart {
                ForEach(days) { day in
                    BarMark(x: .value("Day", day.day, unit: .day), y: .value("Releases", day.pushApprovedCount))
                        .foregroundStyle(by: .value("Status", "Approved"))
                    BarMark(x: .value("Day", day.day, unit: .day), y: .value("Releases", day.pushRejectedCount))
                        .foregroundStyle(by: .value("Status", "Rejected"))
                    BarMark(x: .value("Day", day.day, unit: .day), y: .value("Releases", day.pushErrorCount))
                        .foregroundStyle(by: .value("Status", "Error"))
                }
            }
            .chartForegroundStyleScale(["Approved": Color.green, "Rejected": Color.orange, "Error": Color.red])
            .frame(height: 180)
        }
    }
}

private struct VolumeChart: View {
    let days: [VolumeDay]

    var body: some View {
        if days.allSatisfy({ $0.downloadedBytes == 0 }) {
            EmptyChart()
        } else {
            Chart(days) { day in
                BarMark(x: .value("Day", day.day, unit: .day), y: .value("Bytes", day.downloadedBytes))
                    .foregroundStyle(Color.accentColor)
            }
            .chartYAxis {
                AxisMarks { value in
                    AxisGridLine()
                    AxisValueLabel {
                        if let bytes = value.as(Int64.self) {
                            Text(ByteCountFormatter.string(fromByteCount: bytes, countStyle: .file))
                        }
                    }
                }
            }
            .frame(height: 160)
        }
    }
}

private struct HeatmapChart: View {
    private static let weekdays = Calendar.current.shortWeekdaySymbols

    let cells: [HeatCell]

    var body: some View {
        if cells.allSatisfy({ $0.count == 0 }) {
            EmptyChart()
        } else {
            let peak = Double(cells.map(\.count).max() ?? 1)

            Chart(cells) { cell in
                RectangleMark(
                    x: .value("Hour", cell.hour),
                    y: .value("Day", Self.weekdays[cell.weekday]),
                    width: .ratio(0.9),
                    height: .ratio(0.9)
                )
                .foregroundStyle(Color.accentColor.opacity(cell.count == 0 ? 0.06 : 0.15 + 0.85 * Double(cell.count) / peak))
            }
            .chartXScale(domain: 0...23)
            .chartXAxis {
                AxisMarks(values: [0, 6, 12, 18, 23]) { value in
                    AxisValueLabel { Text("\(value.as(Int.self) ?? 0)h") }
                }
            }
            .chartYScale(domain: Self.weekdays)
            .frame(height: 200)
        }
    }
}

private struct TopList: View {
    let entries: [TopEntry]

    var body: some View {
        if entries.isEmpty {
            EmptyChart()
        } else {
            VStack(spacing: 8) {
                ForEach(entries) { entry in
                    HStack {
                        Text(entry.name).lineLimit(1)
                        Spacer()
                        Text("\(entry.pushApprovedCount.formatted()) / \(entry.matchedCount.formatted())")
                            .font(.callout.monospacedDigit())
                            .foregroundStyle(.secondary)
                    }
                    .font(.callout)
                }

                Text("approved / matched")
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
                    .frame(maxWidth: .infinity, alignment: .trailing)
            }
        }
    }
}

private struct EmptyChart: View {
    var body: some View {
        Text("No data for this period")
            .font(.footnote)
            .foregroundStyle(.secondary)
            .frame(maxWidth: .infinity, minHeight: 60)
    }
}
