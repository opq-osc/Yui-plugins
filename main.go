package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v52/github"
	"github.com/gregjones/httpcache"
	"github.com/opq-osc/Yui/plugin/builder/cmd"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	outPath = "dist"
)

type PluginInfo struct {
	cmd.BuildMetaInfo
	DownloadUrl string
}

func main() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := &http.Client{
		Transport: &oauth2.Transport{
			Base:   httpcache.NewMemoryCacheTransport(),
			Source: ts,
		},
	}
	client := github.NewClient(tc)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "opq-osc", "Yui-plugins")
	if err != nil {
		panic(err)
	}
	dir, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}
	os.MkdirAll(outPath, 0777)
	var pluginsInfo []cmd.BuildMetaInfo
	for _, v := range dir {
		if !v.IsDir() || strings.HasPrefix(v.Name(), ".") {
			continue
		}
		if _, err := os.Stat(filepath.Join(v.Name(), v.Name()+".go")); err != nil {
			continue
		}
		metaBytes, err := os.ReadFile(filepath.Join(v.Name(), "meta.json"))
		if err != nil {
			log.Error(err)
			continue
		}
		var metaInfo = cmd.BuildMetaInfo{}
		err = json.Unmarshal(metaBytes, &metaInfo)
		if err != nil {
			log.Error(err)
			continue
		}
		err = cmd.RunCmd("go", "run", "-mod=mod", "github.com/opq-osc/Yui/plugin/builder", "build", "-o", filepath.Join(outPath, v.Name()), filepath.Join(v.Name(), v.Name()+".go"))
		if err != nil {
			panic(err)
		}
		opqBytes, err := os.ReadFile(filepath.Join(outPath, v.Name()+".opq"))
		if err != nil {
			panic(err)
		}
		sha := sha256.New()
		sha.Write(opqBytes)
		metaInfo.Sha256 = hex.EncodeToString(sha.Sum(nil))
		pluginsInfo = append(pluginsInfo, metaInfo)
		f, err := os.Open(filepath.Join(outPath, v.Name()+".opq"))
		if err != nil {
			panic(err)
		}
		_, _, err = client.Repositories.UploadReleaseAsset(context.Background(), "opq-osc", "Yui-plugins", release.GetID(), &github.UploadOptions{Name: v.Name() + ".opq"}, f)
		if err != nil {
			log.Error(err)
		}
		f.Close()
	}
	pluginsInfoBytes, err := json.Marshal(pluginsInfo)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(outPath, "meta.json"), pluginsInfoBytes, 0777)
	if err != nil {
		panic(err)
	}
	f, err := os.Open(filepath.Join(outPath, "meta.json"))
	if err != nil {
		panic(err)
	}
	_, _, err = client.Repositories.UploadReleaseAsset(context.Background(), "opq-osc", "Yui-plugins", release.GetID(), &github.UploadOptions{Name: "meta.json"}, f)
	if err != nil {
		log.Error(err)
	}
	f.Close()
}
