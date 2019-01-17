package gocbt

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func pullAndStart(t *testing.T, image string) (string, string) {
	cli, err := client.NewEnvClient()
	require.NoError(t, err)

	_, err = cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	require.NoError(t, err)

	return start(t, image)
}

func start(t *testing.T, image string) (string, string) {
	cli, err := client.NewEnvClient()
	require.NoError(t, err)

	ctx := context.Background()
	cont, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
	}, nil, nil, "")
	require.NoError(t, err)

	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	require.NoError(t, err)

	info, err := cli.ContainerInspect(ctx, cont.ID)
	require.NoError(t, err)
	return cont.ID, info.NetworkSettings.IPAddress
}

func stopAndRemove(t *testing.T, id string) {
	cli, err := client.NewEnvClient()
	require.NoError(t, err)

	err = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
	require.NoError(t, err)
}
