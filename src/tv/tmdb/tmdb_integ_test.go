//go:build integ

package tmdb_test

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/igoracmelo/anyflix/src/th"
	"github.com/igoracmelo/anyflix/src/tv"
	"github.com/igoracmelo/anyflix/src/tv/tmdb"
)

func TestFindMovies(t *testing.T) {
	t.Parallel()

	cl := tmdb.DefaultClient()
	cl.DefaultLang = "pt-BR"

	movies, err := cl.FindMovies(context.Background(), tv.FindMoviesParams{
		Title: "mario",
		Page:  1,
	})

	th.Assert.Equal(t, err, nil)
	th.Assert.True(t, len(movies) > 0, "no movies found")

	for _, movie := range movies {
		assertValidContent(t, movie.Content)
	}
}

func TestFindShows(t *testing.T) {
	t.Parallel()

	cl := tmdb.DefaultClient()

	shows, err := cl.FindShows(context.Background(), tv.FindShowsParams{
		Title: "mario",
		Page:  1,
	})

	th.Assert.Equal(t, err, nil)
	th.Assert.True(t, len(shows) > 0, "no shows found")

	for _, show := range shows {
		assertValidContent(t, show.Content)
	}
}

func TestDiscoverMovies(t *testing.T) {
	t.Parallel()

	cl := tmdb.DefaultClient()

	movies, err := cl.DiscoverMovies(context.Background(), tv.DiscoverMoviesParams{
		Page: 1,
	})

	th.Assert.Equal(t, err, nil)
	th.Assert.True(t, len(movies) > 0, "no movies found")

	for _, movie := range movies {
		assertValidContent(t, movie.Content)
	}
}

func TestDiscoverShows(t *testing.T) {
	t.Parallel()

	cl := tmdb.DefaultClient()

	shows, err := cl.DiscoverShows(context.Background(), tv.DiscoverShowsParams{
		Page: 1,
	})

	th.Assert.Equal(t, err, nil)
	th.Assert.True(t, len(shows) > 0, "no shows found")

	for _, show := range shows {
		assertValidContent(t, show.Content)
	}
}

func TestFindMovieDetails(t *testing.T) {
	t.Parallel()

	cl := tmdb.DefaultClient()

	movie, err := cl.FindMovieDetails(context.Background(), "502356")
	th.Assert.Equal(t, err, nil)
	assertValidContent(t, movie.Content)
}

func TestFindShowDetails(t *testing.T) {
	t.Parallel()

	cl := tmdb.DefaultClient()

	show, err := cl.FindShowDetails(context.Background(), "1622")
	th.Assert.Equal(t, err, nil)
	assertValidContent(t, show.Content)
}

func assertValidContent(t *testing.T, content tv.Content) {
	t.Helper()

	_, err := strconv.Atoi(content.ID)
	th.Assert.True(t, content.ID != "", "no id")
	th.Assert.True(t, err == nil, "id is not a valid int: "+content.ID)
	th.Assert.True(t, content.Kind != "", "no kind")
	th.Assert.True(t, content.Title != "", "no title")
	th.Assert.True(t, content.RatingPercent >= -1 && content.RatingPercent <= 100, "invalid rating")

	if content.PosterURL != "" {
		_, err := url.ParseRequestURI(content.PosterURL)
		th.Assert.True(t, err == nil, "invalid URL "+content.PosterURL)
	}
}
