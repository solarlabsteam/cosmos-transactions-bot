package main

import (
	"context"
	"math"

	"google.golang.org/grpc"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type GrpcWrapper struct {
	nodeAddress string
	grpcConn    *grpc.ClientConn
}

func InitGrpcWrapper(nodeAddress string) *GrpcWrapper {
	grpcConn, err := grpc.Dial(
		NodeAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to gRPC node")
	}

	return &GrpcWrapper{
		nodeAddress: nodeAddress,
		grpcConn:    grpcConn,
	}
}

func (w *GrpcWrapper) CloseConnection() {
	grpcConn.Close()
}

func (w *GrpcWrapper) getValidator(address string) (stakingtypes.Validator, error) {
	stakingClient := stakingtypes.NewQueryClient(w.grpcConn)
	validatorResponse, err := stakingClient.Validator(
		context.Background(),
		&stakingtypes.QueryValidatorRequest{ValidatorAddr: address},
	)

	return validatorResponse.Validator, err
}

func (w *GrpcWrapper) setDenom() {
	// if --denom and --denom-coefficient are both provided, use them
	// instead of fetching them via gRPC. Can be useful for networks like osmosis.
	if Denom != "" && DenomCoefficient != 0 {
		log.Info().
			Str("denom", Denom).
			Float64("coefficient", DenomCoefficient).
			Msg("Using provided denom and coefficient.")
		return
	}

	bankClient := banktypes.NewQueryClient(grpcConn)
	denoms, err := bankClient.DenomsMetadata(
		context.Background(),
		&banktypes.QueryDenomsMetadataRequest{},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Error querying denom")
	}

	metadata := denoms.Metadatas[0] // always using the first one
	if Denom == "" {                // using display currency
		Denom = metadata.Display
	}

	for _, unit := range metadata.DenomUnits {
		log.Debug().
			Str("denom", unit.Denom).
			Uint32("exponent", unit.Exponent).
			Msg("Denom info")
		if unit.Denom == Denom {
			DenomCoefficient = math.Pow10(int(unit.Exponent))
			log.Info().
				Str("denom", Denom).
				Float64("coefficient", DenomCoefficient).
				Msg("Got denom info")
			return
		}
	}

	log.Fatal().Msg("Could not find the denom info")
}
