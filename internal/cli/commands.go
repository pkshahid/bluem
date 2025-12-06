package cli

import (
	"context"

	"bluem/internal/deploy"
	"bluem/internal/logx"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var configPath string
	var logLevel string

	root := &cobra.Command{
		Use:   "bluem",
		Short: "Blue-Green Deployment tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logx.Init(logLevel)
		},
	}

	root.PersistentFlags().StringVarP(&configPath, "config", "c", "bluem.yml", "Config file path")
	root.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level: debug|info|warn|error")

	root.AddCommand(
		newInitCmd(&configPath),
		newUpCmd(&configPath),
		newDeployCmd(&configPath),
		newRollbackCmd(&configPath),
		newStatusCmd(&configPath),
		newStopCmd(&configPath),
	)

	return root
}

func newInitCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate compose & proxy configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := deploy.New(ctx, *cfgPath)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Init(ctx)
		},
	}
}

func newUpCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Bring up GREEN stack & proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := deploy.New(ctx, *cfgPath)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Up(ctx)
		},
	}
}

func newDeployCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Perform blue-green deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := deploy.New(ctx, *cfgPath)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Deploy(ctx)
		},
	}
}

func newRollbackCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to GREEN",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := deploy.New(ctx, *cfgPath)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Rollback(ctx)
		},
	}
}

func newStatusCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of stacks & proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := deploy.New(ctx, *cfgPath)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Status(ctx)
		},
	}
}

func newStopCmd(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop blue/green stacks and proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			d, err := deploy.New(ctx, *cfgPath)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Stop(ctx)
		},
	}
}