package playlist

import (
	"context"
)

type Service interface {
	Create(context.Context, *CreatePlaylistCommand) (*Playlist, error)
	Read(context.Context, *GetPlaylistByUidQuery) (*PlaylistDTO, error)
	Update(context.Context, *UpdatePlaylistCommand) (*PlaylistDTO, error)
	Delete(context.Context, *DeletePlaylistCommand) error

	// Non CRUD -- functions
	GetWithoutItems(context.Context, *GetPlaylistByUidQuery) (*Playlist, error)
	Search(context.Context, *GetPlaylistsQuery) (Playlists, error)
}
