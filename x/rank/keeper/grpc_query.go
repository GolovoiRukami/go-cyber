package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	querytypes "github.com/cybercongress/go-cyber/types/query"
	graphtypes "github.com/cybercongress/go-cyber/x/graph/types"
	"github.com/cybercongress/go-cyber/x/rank/types"
)

var _ types.QueryServer = &StateKeeper{}

func (bk StateKeeper) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := bk.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (bk StateKeeper) Rank(goCtx context.Context, req *types.QueryRankRequest) (*types.QueryRankResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNum, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.Cid)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, req.Cid)
	}

	rankValue := bk.index.GetRankValue(cidNum)
	return &types.QueryRankResponse{Rank: rankValue}, nil
}

func (bk *StateKeeper) Search(goCtx context.Context, req *types.QuerySearchRequest) (*types.QuerySearchResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNum, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.Cid)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, "")
	}

	page, limit := uint32(0), uint32(10)
	if req.Pagination != nil {
		page, limit = req.Pagination.Page, req.Pagination.PerPage
	}
	rankedCidNumbers, totalSize, err := bk.index.Search(cidNum, page, limit)
	if err != nil {
		panic(err)
	}

	result := make([]types.RankedCid, 0, len(rankedCidNumbers))
	for _, c := range rankedCidNumbers {
		result = append(result, types.RankedCid{Cid: string(bk.graphKeeper.GetCid(ctx, c.GetNumber())), Rank: c.GetRank()})
	}

	return &types.QuerySearchResponse{Result: result, Pagination: &querytypes.PageResponse{Total: totalSize}}, nil
}

func (bk *StateKeeper) Backlinks(goCtx context.Context, req *types.QuerySearchRequest) (*types.QuerySearchResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNum, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.Cid)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, req.Cid)
	}

	page, limit := uint32(0), uint32(10)
	if req.Pagination != nil {
		page, limit = req.Pagination.Page, req.Pagination.PerPage
	}
	rankedCidNumbers, totalSize, err := bk.index.Backlinks(cidNum, page, limit)
	if err != nil {
		panic(err)
	}

	result := make([]types.RankedCid, 0, len(rankedCidNumbers))
	for _, c := range rankedCidNumbers {
		result = append(result, types.RankedCid{Cid: string(bk.graphKeeper.GetCid(ctx, c.GetNumber())), Rank: c.GetRank()})
	}

	return &types.QuerySearchResponse{Result: result, Pagination: &querytypes.PageResponse{Total: totalSize}}, nil
}

func (bk *StateKeeper) Top(goCtx context.Context, req *querytypes.PageRequest) (*types.QuerySearchResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO check pagination
	page, limit := uint32(0), uint32(100)
	page, limit = req.Page, req.PerPage
	topRankedCidNumbers, totalSize, err := bk.index.Top(page, limit)
	if err != nil {
		panic(err)
	}

	result := make([]types.RankedCid, 0, len(topRankedCidNumbers))
	for _, c := range topRankedCidNumbers {
		result = append(result, types.RankedCid{Cid: string(bk.graphKeeper.GetCid(ctx, c.GetNumber())), Rank: c.GetRank()})
	}

	return &types.QuerySearchResponse{Result: result, Pagination: &querytypes.PageResponse{Total: totalSize}}, nil
}

func (bk StateKeeper) IsLinkExist(goCtx context.Context, req *types.QueryIsLinkExistRequest) (*types.QueryLinkExistResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address); if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNumFrom, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.From)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, req.From)
	}

	cidNumTo, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.To)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, req.To)
	}

	var accountNum uint64
	account := bk.accountKeeper.GetAccount(ctx, addr)
	if account != nil {
		accountNum = account.GetAccountNumber()
	} else {
		return nil, sdkerrors.Wrap(graphtypes.ErrInvalidAccount, addr.String())
	}

	exists := bk.graphIndexedKeeper.IsLinkExist(graphtypes.CompactLink{
		uint64(cidNumFrom),
		uint64(cidNumTo),
		accountNum,
	})

	return &types.QueryLinkExistResponse{Exist: exists}, nil
}

func (bk StateKeeper) IsAnyLinkExist(goCtx context.Context, req *types.QueryIsAnyLinkExistRequest) (*types.QueryLinkExistResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNumFrom, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.From)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, req.From)
	}

	cidNumTo, exist := bk.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(req.To)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, req.To)
	}

	exists := bk.graphIndexedKeeper.IsAnyLinkExist(cidNumFrom, cidNumTo)

	return &types.QueryLinkExistResponse{Exist: exists}, nil
}

func (s *StateKeeper) Entropy(goCtx context.Context, request *types.QueryEntropyRequest) (*types.QueryEntropyResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNum, exist := s.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(request.Cid)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, request.Cid)
	}

	entropyValue := s.GetEntropy(cidNum)
	return &types.QueryEntropyResponse{Entropy: entropyValue}, nil
}

func (s *StateKeeper) Luminosity(goCtx context.Context, request *types.QueryLuminosityRequest) (*types.QueryLuminosityResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	cidNum, exist := s.graphKeeper.GetCidNumber(ctx, graphtypes.Cid(request.Cid)); if exist != true {
		return nil, sdkerrors.Wrap(graphtypes.ErrCidNotFound, request.Cid)
	}

	luminosityValue := s.GetLuminosity(cidNum)
	return &types.QueryLuminosityResponse{Luminosity: luminosityValue}, nil
}

func (s *StateKeeper) Karma(goCtx context.Context, request *types.QueryKarmaRequest) (*types.QueryKarmaResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(request.Address); if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var accountNum uint64
	account := s.accountKeeper.GetAccount(ctx, addr)
	if account != nil {
		accountNum = account.GetAccountNumber()
	} else {
		return nil, sdkerrors.Wrap(graphtypes.ErrInvalidAccount, addr.String())
	}

	karma := s.GetKarma(accountNum)

	return &types.QueryKarmaResponse{Karma: karma}, nil
}