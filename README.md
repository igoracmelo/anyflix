# Anyflix

[WIP] A self-hosted streaming application for movies and tv shows.

![photo_2023-12-30_01-56-00](https://github.com/igoracmelo/anyflix/assets/85039990/61953c96-3c8a-4510-aa80-82ae4625dcd1)

# Running

**Required:** Go +1.21, Git, MPV, Web browser

1. Install the dependencies

It depends on your system, so please do your own research.

2. Clone the repository

```sh
git clone https://github.com/igoracmelo/anyflix
```

3. Run the server

```sh
go run .
```

4. Open the browser at `http://localhost:3000`

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
- [ ] Embed all web assets on binary
- [ ] Choose prefered language for content
- [ ] Previous page button
- [ ] Content page responsivity
- [ ] Offline
    - [ ] Favorite shows
    - [ ] Favorite movies
