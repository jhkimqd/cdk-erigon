package txpool

import (
	"context"
	"testing"

	"github.com/ledgerwatch/erigon-lib/direct"
	"github.com/ledgerwatch/erigon-lib/gointerfaces"
	"github.com/ledgerwatch/erigon-lib/gointerfaces/sentry"
	"github.com/ledgerwatch/erigon-lib/gointerfaces/types"
	"github.com/ledgerwatch/erigon-lib/txpool"
	types2 "github.com/ledgerwatch/erigon-lib/types"
	"github.com/ledgerwatch/log/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
)

func testRlps(num int) [][]byte {
	rlps := make([][]byte, num)
	for i := 0; i < num; i++ {
		rlps[i] = []byte{1}
	}
	return rlps
}

func toHashes(h ...byte) (out types2.Hashes) {
	for i := range h {
		hash := [32]byte{h[i]}
		out = append(out, hash[:]...)
	}
	return out
}

func toPeerIDs(h ...byte) (out []types2.PeerID) {
	for i := range h {
		hash := [64]byte{h[i]}
		out = append(out, gointerfaces.ConvertHashToH512(hash))
	}
	return out
}

func TestSendTxPropagate(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	t.Run("few remote byHash", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		sentryServer := sentry.NewMockSentryServer(ctrl)
		times := 2
		requests := make([]*sentry.SendMessageToRandomPeersRequest, 0, times)
		sentryServer.EXPECT().
			SendMessageToRandomPeers(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *sentry.SendMessageToRandomPeersRequest) (*sentry.SentPeers, error) {
				requests = append(requests, r)
				return nil, nil
			}).
			Times(times)

		sentryServer.EXPECT().PeerById(gomock.Any(), gomock.Any()).
			DoAndReturn(
				func(_ context.Context, r *sentry.PeerByIdRequest) (*sentry.PeerByIdReply, error) {
					return &sentry.PeerByIdReply{
						Peer: &types.PeerInfo{
							Id:   r.PeerId.String(),
							Caps: []string{"eth/68"},
						}}, nil
				}).AnyTimes()

		m := txpool.NewMockSentry(ctx, sentryServer)
		send := txpool.NewSend(ctx, []direct.SentryClient{direct.NewSentryClientDirect(direct.ETH68, m)}, nil, log.New())
		send.BroadcastPooledTxs(testRlps(2), 100)
		send.AnnouncePooledTxs([]byte{0, 1}, []uint32{10, 15}, toHashes(1, 42), 100)

		require.Equal(t, 2, len(requests))

		txsMessage := requests[0].Data
		assert.Equal(t, sentry.MessageId_TRANSACTIONS_66, txsMessage.Id)
		assert.Equal(t, 3, len(txsMessage.Data))

		txnHashesMessage := requests[1].Data
		assert.Equal(t, sentry.MessageId_NEW_POOLED_TRANSACTION_HASHES_68, txnHashesMessage.Id)
		assert.Equal(t, 76, len(txnHashesMessage.Data))
	})

	// t.Run("much remote byHash", func(t *testing.T) {
	// 	ctrl := gomock.NewController(t)
	// 	sentryServer := sentry.NewMockSentryServer(ctrl)

	// 	times := 2
	// 	requests := make([]*sentry.SendMessageToRandomPeersRequest, 0, times)
	// 	sentryServer.EXPECT().
	// 		SendMessageToRandomPeers(gomock.Any(), gomock.Any()).
	// 		DoAndReturn(func(_ context.Context, r *sentry.SendMessageToRandomPeersRequest) (*sentry.SentPeers, error) {
	// 			requests = append(requests, r)
	// 			return nil, nil
	// 		}).
	// 		Times(times)

	// 	m := txpool.NewMockSentry(ctx, sentryServer)
	// 	send := NewSend(ctx, []direct.SentryClient{direct.NewSentryClientDirect(direct.ETH68, m)}, nil)
	// 	list := make(types2.Hashes, p2pTxPacketLimit*3)
	// 	for i := 0; i < len(list); i += 32 {
	// 		b := []byte(fmt.Sprintf("%x", i))
	// 		copy(list[i:i+32], b)
	// 	}
	// 	send.BroadcastPooledTxs(testRlps(len(list) / 32))
	// 	send.AnnouncePooledTxs([]byte{0, 1, 2}, []uint32{10, 12, 14}, list)

	// 	require.Equal(t, 2, len(requests))

	// 	txsMessage := requests[0].Data
	// 	require.Equal(t, sentry.MessageId_TRANSACTIONS_66, txsMessage.Id)
	// 	require.True(t, len(txsMessage.Data) > 0)

	// 	txnHashesMessage := requests[1].Data
	// 	require.Equal(t, sentry.MessageId_NEW_POOLED_TRANSACTION_HASHES_68, txnHashesMessage.Id)
	// 	require.True(t, len(txnHashesMessage.Data) > 0)
	// })

	// t.Run("few local byHash", func(t *testing.T) {
	// 	ctrl := gomock.NewController(t)
	// 	sentryServer := sentry.NewMockSentryServer(ctrl)

	// 	times := 2
	// 	requests := make([]*sentry.SendMessageToRandomPeersRequest, 0, times)
	// 	sentryServer.EXPECT().
	// 		SendMessageToRandomPeers(gomock.Any(), gomock.Any()).
	// 		DoAndReturn(func(_ context.Context, r *sentry.SendMessageToRandomPeersRequest) (*sentry.SentPeers, error) {
	// 			requests = append(requests, r)
	// 			return nil, nil
	// 		}).
	// 		Times(times)

	// 	m := txpool.NewMockSentry(ctx, sentryServer)
	// 	send := NewSend(ctx, []direct.SentryClient{direct.NewSentryClientDirect(direct.ETH68, m)}, nil)
	// 	send.BroadcastPooledTxs(testRlps(2))
	// 	send.AnnouncePooledTxs([]byte{0, 1}, []uint32{10, 15}, toHashes(1, 42))

	// 	require.Equal(t, 2, len(requests))

	// 	txsMessage := requests[0].Data
	// 	assert.Equal(t, sentry.MessageId_TRANSACTIONS_66, txsMessage.Id)
	// 	assert.True(t, len(txsMessage.Data) > 0)

	// 	txnHashesMessage := requests[1].Data
	// 	assert.Equal(t, sentry.MessageId_NEW_POOLED_TRANSACTION_HASHES_68, txnHashesMessage.Id)
	// 	assert.Equal(t, 76, len(txnHashesMessage.Data))
	// })

	t.Run("sync with new peer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		sentryServer := sentry.NewMockSentryServer(ctrl)

		times := 3
		requests := make([]*sentry.SendMessageByIdRequest, 0, times)
		sentryServer.EXPECT().
			SendMessageById(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *sentry.SendMessageByIdRequest) (*sentry.SentPeers, error) {
				requests = append(requests, r)
				return nil, nil
			}).
			Times(times)

		sentryServer.EXPECT().PeerById(gomock.Any(), gomock.Any()).
			DoAndReturn(
				func(_ context.Context, r *sentry.PeerByIdRequest) (*sentry.PeerByIdReply, error) {
					return &sentry.PeerByIdReply{
						Peer: &types.PeerInfo{
							Id:   r.PeerId.String(),
							Caps: []string{"eth/68"},
						}}, nil
				}).AnyTimes()

		m := txpool.NewMockSentry(ctx, sentryServer)
		send := NewSend(ctx, []direct.SentryClient{direct.NewSentryClientDirect(direct.ETH68, m)}, nil)
		expectPeers := toPeerIDs(1, 2, 42)
		send.PropagatePooledTxsToPeersList(expectPeers, []byte{0, 1}, []uint32{10, 15}, toHashes(1, 42))

		require.Equal(t, 3, len(requests))
		for i, req := range requests {
			assert.Equal(t, expectPeers[i], types2.PeerID(req.PeerId))
			assert.Equal(t, sentry.MessageId_NEW_POOLED_TRANSACTION_HASHES_68, req.Data.Id)
			assert.True(t, len(req.Data.Data) > 0)
		}
	})
}
