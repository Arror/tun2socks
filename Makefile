build:
	gomobile bind -o ./MoonKit.xcframework -target=macos -ldflags="-s -w" -v ./