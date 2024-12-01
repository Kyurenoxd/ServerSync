package main

type CloneStats struct {
	Roles         int
	Categories    int
	TextChannels  int
	VoiceChannels int
}

type Guild struct {
	ID       string
	Name     string
	IconURL  string
	Roles    []Role
	Channels []Channel
}

type Role struct {
	ID          string
	Name        string
	Color       int
	Hoist       bool
	Permissions int64
	Mentionable bool
	Position    int
}

type Channel struct {
	ID        string
	Name      string
	Type      string
	ParentID  string
	Topic     string
	NSFW      bool
	Bitrate   int
	UserLimit int
}
