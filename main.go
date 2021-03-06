package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/int128/gpup/oauth"
	"github.com/int128/gpup/photos"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	NewAlbum     string `short:"n" long:"new-album" value-name:"TITLE" description:"Create an album and add files into it"`
	OAuthMethod  string `long:"oauth-method" default:"browser" choice:"browser" choice:"cli" description:"OAuth authorization method"`
	ClientID     string `long:"google-client-id" env:"GOOGLE_CLIENT_ID" required:"1" description:"Google API client ID"`
	ClientSecret string `long:"google-client-secret" env:"GOOGLE_CLIENT_SECRET" required:"1" description:"Google API client secret"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] FILE or DIRECTORY..."
	parser.LongDescription = oauthDescription
	args, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	files, err := findFiles(args)
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatalf("File not found in %s", strings.Join(args, ", "))
	}
	log.Printf("The following %d files will be uploaded:", len(files))
	for i, file := range files {
		fmt.Printf("%3d: %s\n", i+1, file)
	}

	ctx := context.Background()
	client, err := oauth.NewClient(ctx, opts.OAuthMethod, opts.ClientID, opts.ClientSecret)
	if err != nil {
		log.Fatal(err)
	}
	service, err := photos.New(client)
	if err != nil {
		log.Fatal(err)
	}

	if opts.NewAlbum != "" {
		_, err := service.CreateAlbum(ctx, opts.NewAlbum, files)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if err := service.AddToLibrary(ctx, files); err != nil {
			log.Fatal(err)
		}
	}
}

func findFiles(filePaths []string) ([]string, error) {
	files := make([]string, 0, len(filePaths)*2)
	for _, parent := range filePaths {
		if err := filepath.Walk(parent, func(child string, info os.FileInfo, err error) error {
			switch {
			case err != nil:
				return err
			case info.Mode().IsRegular():
				files = append(files, child)
				return nil
			default:
				return nil
			}
		}); err != nil {
			return nil, fmt.Errorf("Error while finding files in %s: %s", parent, err)
		}
	}
	return files, nil
}

const oauthDescription = `
	Setup:
	1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
	2. Enable Photos Library API.
	3. Open https://console.cloud.google.com/apis/credentials
	4. Create an OAuth client ID where the application type is other.
	5. Export GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET variables or set the options.`
