# autobrr for iOS

A native SwiftUI app around the autobrr engine. The Go engine is built into `AutobrrEngine.xcframework` with gomobile and runs inside the app. Screens call its API in-process through `EngineClient`, so nothing listens on the network.

## Requirements

- macOS with Xcode 15 or later
- Go (the version in `go.mod`)
- [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile): `go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init`
- [XcodeGen](https://github.com/yonaskolb/XcodeGen): `brew install xcodegen`

## Build

From the repository root:

```sh
make build/ios                       # bin/AutobrrEngine.xcframework
cd ios && xcodegen generate          # writes Autobrr.xcodeproj
open Autobrr.xcodeproj
```

Set your team under Signing & Capabilities (and change `com.autobrr.ios` if it is taken), then run on a device or simulator. Re-run `make build/ios` after changing Go code.

## Sideloading

Archive with Product → Archive and export an `.ipa` for Ad Hoc or Development, or install directly from Xcode. A free Apple ID signs apps for 7 days.

## Limitations

- iOS suspends the app shortly after it leaves the foreground, so IRC announces and feeds are only processed while it is open.
- Data lives in the app's Application Support folder and is removed with the app.
