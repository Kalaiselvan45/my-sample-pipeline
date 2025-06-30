package impl

import (
	"fmt"
	"log/slog"
)

type DockerCmd struct {
	CreateManifest DockerCreateManifestCmd `cmd:"" help:"docker create-manifest command is to create a manifest image"`
}

type DockerCreateManifestCmd struct {
	Repos         []string `help:"list repo for publishing the image manifest" env:"PUBLISH_REPOS"`
	ImageTag      string   `help:"image-tag to be applied" env:"IMAGE_TAG" required:""`
	AdditionalTag string   `help:"Additional tag name" default:"latest"`
}

// This can be made to run in parallel as well
func createManifest(cli CLI) error {
	repos := cli.Docker.CreateManifest.Repos
	if len(repos) == 0 {
		return fmt.Errorf("image repository not found")
	}
	imgTag := cli.Docker.CreateManifest.ImageTag
	tags := []string{imgTag}
	tags = append(tags, cli.Docker.CreateManifest.AdditionalTag)
	archs := []string{"amd64", "arm64"}
	slog.Info("Creating and pushing multi-arch manifest...", "repos", repos, "imgTag", imgTag)
	for _, tag := range tags {
		if err := dockerManifestCreate(repos, tag); err != nil {
			return err
		}
		for _, arch := range archs {
			if err := dockerManifestAnnotate(repos, tag, arch); err != nil {
				return err
			}
		}
		if err := dockerManifestPush(repos, tag); err != nil {
			return err
		}
	}
	slog.Info("Manifest created successfully.")
	return nil
}

// dockerManifestCreate creates a multi-architecture manifest
func dockerManifestCreate(repos []string, tag string) error {
	for _, repo := range repos {
		err := run("docker", "manifest", "create",
			fmt.Sprintf("%s:%s", repo, tag),
			fmt.Sprintf("%s:%s-amd64", repo, tag),
			fmt.Sprintf("%s:%s-arm64", repo, tag),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// dockerManifestAnnotate sets architecture for a specific manifest entry
func dockerManifestAnnotate(repos []string, tag, arch string) error {
	for _, repo := range repos {
		err := run("docker", "manifest", "annotate",
			fmt.Sprintf("%s:%s", repo, tag),
			fmt.Sprintf("%s:%s-%s", repo, tag, arch),
			"--arch", arch,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// dockerManifestPush pushes the manifest to the repository
func dockerManifestPush(repos []string, tag string) error {
	for _, repo := range repos {
		err := run("docker", "manifest", "push", fmt.Sprintf("%s:%s", repo, tag))
		if err != nil {
			return err
		}
	}
	return nil
}
