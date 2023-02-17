package relay

import (
	"encoding/hex"
	"github.com/datachainlab/lcp/go/relay/elc"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagSrc        = "src"
	flagCommitment = "commitment"
)

func LCPCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lcp",
		Short: "LCP commands",
	}

	cmd.AddCommand(
		registerEnclaveKeyCmd(ctx),
		activateClientCmd(ctx),
	)

	return cmd
}

func registerEnclaveKeyCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-key [path]",
		Short: "Register an enclave key into the LCP client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, src, dst, err := ctx.Config.ChainsFromPath(args[0])
			if err != nil {
				return err
			}
			path, err := ctx.Config.Paths.Get(args[0])
			if err != nil {
				return err
			}
			var (
				pathEnd *core.PathEnd
				target  *core.ProvableChain
			)
			if viper.GetBool(flagSrc) {
				pathEnd = path.Src
				target = c[src]
			} else {
				pathEnd = path.Dst
				target = c[dst]
			}
			prover := target.ProverI.(*Prover)
			if err := prover.initServiceClient(); err != nil {
				return err
			}
			// TODO add debug option
			return registerEnclaveKey(pathEnd, prover, true)
		},
	}
	return srcFlag(cmd)
}

func activateClientCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-client [path]",
		Short: "Activate the LCP client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, src, dst, err := ctx.Config.ChainsFromPath(args[0])
			if err != nil {
				return err
			}
			path, err := ctx.Config.Paths.Get(args[0])
			if err != nil {
				return err
			}
			var (
				pathEnd      *core.PathEnd
				target       *core.ProvableChain
				counterparty *core.ProvableChain
			)
			if viper.GetBool(flagSrc) {
				pathEnd = path.Src
				target, counterparty = c[src], c[dst]
			} else {
				pathEnd = path.Dst
				target, counterparty = c[dst], c[src]
			}
			commitment := viper.GetString(flagCommitment)
			var msgRes *elc.MsgUpdateClientResponse = nil
			if commitment != "" {
				commitmentBytes, err := hex.DecodeString(commitment)
				if err != nil {
					return err
				}
				createClient := &elc.MsgCreateClientResponse{}
				err = createClient.Unmarshal(commitmentBytes)
				if err != nil {
					return err
				}
				msgRes = &elc.MsgUpdateClientResponse{
					Commitment: createClient.Commitment,
					Signer:     createClient.Signer,
					Signature:  createClient.Signature,
				}
			}
			return activateClient(pathEnd, target, counterparty, msgRes)
		},
	}

	cmd.Flags().String(flagCommitment, "", "a hexadecimal value of MsgCreateClientResponse. This flag skips a UpdateClient.")

	return srcFlag(cmd)
}

func srcFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().BoolP(flagSrc, "", true, "a boolean value whether src is the target chain")
	if err := viper.BindPFlag(flagSrc, cmd.Flags().Lookup(flagSrc)); err != nil {
		panic(err)
	}
	return cmd
}
