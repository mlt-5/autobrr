// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SafariServices
import SwiftUI
import WebKit

/// WebView shows the autobrr web UI and opens links to other sites in Safari instead of navigating away from it.
struct WebView: UIViewRepresentable {
    let handler: SchemeHandler

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    func makeUIView(context: Context) -> WKWebView {
        let configuration = WKWebViewConfiguration()
        configuration.setURLSchemeHandler(handler, forURLScheme: SchemeHandler.scheme)

        let webView = WKWebView(frame: .zero, configuration: configuration)
        webView.navigationDelegate = context.coordinator
        webView.uiDelegate = context.coordinator
        webView.allowsBackForwardNavigationGestures = true
        #if DEBUG
        webView.isInspectable = true
        #endif

        webView.load(URLRequest(url: SchemeHandler.startURL))

        return webView
    }

    func updateUIView(_ webView: WKWebView, context: Context) {}

    final class Coordinator: NSObject, WKNavigationDelegate, WKUIDelegate {
        private static let internalSchemes: Set<String> = [SchemeHandler.scheme, "about", "blob", "data"]

        func webView(
            _ webView: WKWebView,
            decidePolicyFor navigationAction: WKNavigationAction,
            decisionHandler: @escaping (WKNavigationActionPolicy) -> Void
        ) {
            guard let url = navigationAction.request.url, !Self.internalSchemes.contains(url.scheme ?? "") else {
                decisionHandler(.allow)
                return
            }

            decisionHandler(.cancel)
            openExternally(url, from: webView)
        }

        func webView(
            _ webView: WKWebView,
            createWebViewWith configuration: WKWebViewConfiguration,
            for navigationAction: WKNavigationAction,
            windowFeatures: WKWindowFeatures
        ) -> WKWebView? {
            if let url = navigationAction.request.url {
                openExternally(url, from: webView)
            }

            return nil
        }

        private func openExternally(_ url: URL, from webView: WKWebView) {
            guard url.scheme == "http" || url.scheme == "https" else {
                UIApplication.shared.open(url)
                return
            }

            webView.window?.rootViewController?.present(SFSafariViewController(url: url), animated: true)
        }
    }
}
