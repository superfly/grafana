package playlisttest

import (
	"context"

	"github.com/grafana/grafana/pkg/services/playlist"
)

type FakePlaylistService struct {
	ExpectedPlaylist      *playlist.Playlist
	ExpectedPlaylistDTO   *playlist.PlaylistDTO
	ExpectedPlaylistItems []playlist.PlaylistItem
	ExpectedPlaylists     playlist.Playlists
	ExpectedError         error
}

// Make sure it implements the service
var _ playlist.Service = &FakePlaylistService{}

func NewPlaylistServiveFake() *FakePlaylistService {
	return &FakePlaylistService{}
}

func (f *FakePlaylistService) Create(context.Context, *playlist.CreatePlaylistCommand) (*playlist.Playlist, error) {
	return f.ExpectedPlaylist, f.ExpectedError
}

func (f *FakePlaylistService) Read(context.Context, *playlist.GetPlaylistByUidQuery) (*playlist.PlaylistDTO, error) {
	return f.ExpectedPlaylistDTO, f.ExpectedError
}

func (f *FakePlaylistService) Update(context.Context, *playlist.UpdatePlaylistCommand) (*playlist.PlaylistDTO, error) {
	return f.ExpectedPlaylistDTO, f.ExpectedError
}

func (f *FakePlaylistService) GetWithoutItems(context.Context, *playlist.GetPlaylistByUidQuery) (*playlist.Playlist, error) {
	return f.ExpectedPlaylist, f.ExpectedError
}

func (f *FakePlaylistService) Search(context.Context, *playlist.GetPlaylistsQuery) (playlist.Playlists, error) {
	return f.ExpectedPlaylists, f.ExpectedError
}

func (f *FakePlaylistService) Delete(ctx context.Context, cmd *playlist.DeletePlaylistCommand) error {
	return f.ExpectedError
}
