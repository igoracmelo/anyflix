# Anyflix

[WIP] A self-hosted streaming application for movies and tv shows.

![photo_2023-12-31_17-06-20](https://github.com/igoracmelo/anyflix/assets/85039990/350df101-5afc-4882-a1bd-edacd9ef8c00)

# Running

**Required:** Go +1.21, web browser

**Optional:** MPV

### 1. Install the dependencies

It varies by system, so please do your own research.

### 2. Install anyflix using `go install`

```sh
go install https://github.com/igoracmelo/anyflix@b00cecf
# I don't recommend using @latest because it is usually outdated
```

### 3. Run the application

```sh
$(go env GOPATH)/bin/anyflix
```

Which is usually equivalent to

```sh
~/go/bin/anyflix
```

Or even better if `~/go/bin/` is in your `PATH`

```sh
anyflix
```

### 4. Open your browser and visit `http://localhost:3000`

# Known issues

Some audio and video codecs are not supported on browsers, but most will work on MPV.
I'm trying to find a way to transcode it "on demand" without needing HLS or similar solutions, but I think I will end up needing it, or just giving up on supporting these codecs and force you to use MPV.

# TODO
- [X] stream torrent video using HTTP range requests
- [X] play video on MPV
- [ ] Movies
    - [X] Get specific movie info
    - [X] List popular movies
    - [X] Search movies
    - [X] Find torrents for a specific movie
    - [X] Guess resolution
    - [ ] Movie details (ratings, casting, original title, original language, trailer, duration)
- [ ] Include embeded subtitles in web player
- [ ] Include embeded audio tracks in web player
- [ ] Replace mpv executable with libmpv
- [X] Embed pages on binary
- [X] Embed all web assets on binary
- [ ] Choose prefered language for content
- [ ] Previous page button
- [X] Content page responsivity
- [ ] Offline
    - [ ] Favorite shows
    - [ ] Favorite movies
