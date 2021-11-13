package messages

import (
	"context"
	"fmt"

	"go.elastic.co/apm"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/messages/create"
	"mealswipe.app/mealswipe/internal/messages/join"
	"mealswipe.app/mealswipe/internal/messages/rejoin"
	"mealswipe.app/mealswipe/internal/messages/start"
	"mealswipe.app/mealswipe/internal/messages/vote"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO NULL SAFETY FROM PROTOBUF STUFF
func HandleMessage(ctx context.Context, userState *types.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	logger := logging.Get()
	if common.HasCreateMessage(genericMessage) {
		tx := apm.DefaultTracer.StartTransaction("HANDLE create", "request")
		defer tx.End()
		ctx = apm.ContextWithTransaction(ctx, tx)

		logger.Info(fmt.Sprintf("create message received from %s", userState.UserId),
			logging.Metric("message_received"),
			zap.String("type", "create"),
		)
		err = create.HandleMessage(ctx, userState, genericMessage.GetCreateMessage())
		if err != nil {
			err = fmt.Errorf("handle create message: %w", err)
		}
		return
	} else if common.HasJoinMessage(genericMessage) {
		tx := apm.DefaultTracer.StartTransaction("HANDLE join", "request")
		defer tx.End()
		ctx = apm.ContextWithTransaction(ctx, tx)

		logger.Info(fmt.Sprintf("join message received from %s", userState.UserId),
			logging.Metric("message_received"),
			zap.String("type", "join"),
		)
		err = join.HandleMessage(ctx, userState, genericMessage.GetJoinMessage())
		if err != nil {
			err = fmt.Errorf("handle join message: %w", err)
		}
		return
	} else if common.HasStartMessage(genericMessage) {
		tx := apm.DefaultTracer.StartTransaction("HANDLE start", "request")
		defer tx.End()
		ctx = apm.ContextWithTransaction(ctx, tx)

		logger.Info(fmt.Sprintf("start message received from %s", userState.UserId),
			logging.Metric("message_received"),
			zap.String("type", "start"),
		)
		err = start.HandleMessage(ctx, userState, genericMessage.GetStartMessage())
		if err != nil {
			err = fmt.Errorf("handle start message: %w", err)
		}
		return
	} else if common.HasVoteMessage(genericMessage) {
		tx := apm.DefaultTracer.StartTransaction("HANDLE vote", "request")
		defer tx.End()
		ctx = apm.ContextWithTransaction(ctx, tx)

		logger.Info(fmt.Sprintf("vote message received from %s", userState.UserId),
			logging.Metric("message_received"),
			zap.String("type", "vote"),
		)
		err = vote.HandleMessage(ctx, userState, genericMessage.GetVoteMessage())
		if err != nil {
			err = fmt.Errorf("handle vote message: %w", err)
		}
		return
	} else if common.HasRejoinMessage(genericMessage) {
		tx := apm.DefaultTracer.StartTransaction("HANDLE rejoin", "request")
		defer tx.End()
		ctx = apm.ContextWithTransaction(ctx, tx)

		logger.Info(fmt.Sprintf("rejoin message received from %s", userState.UserId),
			logging.Metric("message_received"),
			zap.String("type", "rejoin"),
		)
		err = rejoin.HandleMessage(ctx, userState, genericMessage.GetRejoinMessage())
		if err != nil {
			err = fmt.Errorf("handle rejoin message: %w", err)
		}
		return
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
