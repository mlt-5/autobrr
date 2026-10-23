// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import CoreLocation
import Foundation

/// BackgroundKeepAlive keeps the app, and with it the engine, running in the background by receiving location updates.
///
/// iOS suspends an app seconds after it leaves the foreground, which drops IRC connections and stops feeds. An app
/// receiving location updates in the background is not suspended. The locations are discarded; accuracy and distance
/// filter are set as coarse as possible so the GPS stays off.
@MainActor
final class BackgroundKeepAlive: NSObject, ObservableObject {
    private static let enabledKey = "keepAliveInBackground"

    @Published private(set) var isEnabled: Bool
    @Published private(set) var authorization: CLAuthorizationStatus
    @Published private(set) var isRunning = false

    private let manager = CLLocationManager()

    override init() {
        isEnabled = UserDefaults.standard.bool(forKey: Self.enabledKey)
        authorization = manager.authorizationStatus

        super.init()

        manager.delegate = self
        manager.desiredAccuracy = kCLLocationAccuracyThreeKilometers
        manager.distanceFilter = CLLocationDistanceMax
        manager.activityType = .other
        // Paused updates let iOS suspend the app, which is what this is here to prevent.
        manager.pausesLocationUpdatesAutomatically = false
    }

    var isDenied: Bool {
        authorization == .denied || authorization == .restricted
    }

    func setEnabled(_ enabled: Bool) {
        isEnabled = enabled
        UserDefaults.standard.set(enabled, forKey: Self.enabledKey)
        update()
    }

    /// update starts or stops location updates to match the setting and authorization. Updates must start while the
    /// app is in the foreground, so this runs at launch and whenever the setting or authorization changes.
    func update() {
        guard isEnabled else {
            stop()
            return
        }

        switch authorization {
        case .notDetermined:
            manager.requestWhenInUseAuthorization()
        case .authorizedWhenInUse, .authorizedAlways:
            start()
        default:
            stop()
        }
    }

    private func start() {
        guard !isRunning else { return }

        manager.allowsBackgroundLocationUpdates = true
        manager.showsBackgroundLocationIndicator = true
        manager.startUpdatingLocation()
        isRunning = true
    }

    private func stop() {
        guard isRunning else { return }

        manager.stopUpdatingLocation()
        manager.allowsBackgroundLocationUpdates = false
        isRunning = false
    }
}

extension BackgroundKeepAlive: CLLocationManagerDelegate {
    nonisolated func locationManagerDidChangeAuthorization(_ manager: CLLocationManager) {
        let status = manager.authorizationStatus

        Task { @MainActor in
            self.authorization = status
            self.update()
        }
    }

    nonisolated func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {}

    nonisolated func locationManager(_ manager: CLLocationManager, didFailWithError error: Error) {}
}
