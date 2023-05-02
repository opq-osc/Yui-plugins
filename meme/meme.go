//go:build tinygo.wasm

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/knqyf263/go-plugin/types/known/emptypb"
	"github.com/opq-osc/OPQBot/v2/events"
	"github.com/opq-osc/Yui-plugins/meme/model"
	"github.com/opq-osc/Yui/plugin/S"
	"github.com/opq-osc/Yui/proto"
	"mime/multipart"
	"os"
	"strings"
)

var Keys map[string]*model.KeyInfoRes

type meme struct {
}

func GetQQPic(ctx context.Context, Uin int64) ([]byte, error) {
	res, err := S.HttpGet(ctx, fmt.Sprintf("https://q.qlogo.cn/g?b=qq&nk=%v&s=100", Uin), nil)
	if err != nil {
		return nil, err
	}
	return res.Content, err
}

func (p meme) OnRemoteCallEvent(ctx context.Context, req *proto.RemoteCallReq) (*proto.RemoteCallReply, error) {
	//TODO implement me
	panic("implement me")
}

func (p meme) OnCronEvent(ctx context.Context, req *proto.CronEventReq) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (p meme) OnFriendMsg(ctx context.Context, msg *proto.CommonMsg) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (p meme) OnPrivateMsg(ctx context.Context, msg *proto.CommonMsg) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (p meme) OnGroupMsg(ctx context.Context, msg *proto.CommonMsg) (*emptypb.Empty, error) {
	event, err := S.ParserEvent(msg.RawMessage)
	if err != nil {
		S.LogError(ctx, err.Error())
		return nil, err
	}
	if event.ParseGroupMsg().ParseTextMsg().GetTextContent() == ".meme help" {
		help := []string{"meme 支持的有"}
		for k, _ := range Keys {
			help = append(help, k)
		}
		S.SendGroupTextMsg(ctx, event.ParseGroupMsg().GetGroupUin(), event.GetCurrentQQ(), strings.Join(help, "\n"))
	}
	if event.ParseGroupMsg().ContainedAt() {
		if v, ok := Keys[event.ParseGroupMsg().ExcludeAtInfo().ParseTextMsg().GetTextContent()]; ok {
			var buf bytes.Buffer
			m := multipart.NewWriter(&buf)
			atInfo := event.ParseGroupMsg().GetAtInfo()
			users := append(atInfo, events.UserInfo{
				Uin:  event.ParseGroupMsg().GetSenderUin(),
				Nick: event.ParseGroupMsg().GetSenderNick(),
			})
			if len(users) < v.Params.MinImages {
				S.SendGroupTextMsg(ctx, event.ParseGroupMsg().GetGroupUin(), event.GetCurrentQQ(), fmt.Sprintf("您需要at至少%d个人", v.Params.MinImages-1))
				return nil, nil
			}
			for i := 0; i < v.Params.MaxImages; i++ {
				img, _ := m.CreateFormFile("images", "pic1.jpg")
				pic, _ := GetQQPic(ctx, users[i].Uin)
				img.Write(pic)
			}
			b, _ := m.CreateFormField("texts")
			texts := []string{}
			for i := 0; i < v.Params.MaxTexts; i++ {
				texts = append(texts, users[i].Nick)
			}
			b.Write([]byte(strings.Join(texts, ",")))
			m.Close()
			res, err := S.HttpPost(ctx, fmt.Sprintf(os.Getenv("memeUrl")+"/memes/%s/", v.Key), map[string]string{"accept": "accept", "Content-Type": m.FormDataContentType()}, buf.Bytes())
			if err != nil {
				S.LogError(ctx, err.Error())
				return nil, err
			}
			file, err := proto.NewApi().Upload(ctx, &proto.UploadReq{
				File: &proto.UploadReq_Base64Buf{
					Base64Buf: base64.StdEncoding.EncodeToString(res.Content),
				},
				BotUin:   event.GetCurrentQQ(),
				UploadId: proto.UploadId_Group,
			})
			if err != nil {
				S.LogError(ctx, err.Error())
				return nil, err
			}
			//S.LogInfo(ctx, fmt.Sprintf("%v", file))
			_, err = proto.NewApi().SendGroupMsg(ctx, &proto.MsgReq{
				ToUin:  event.ParseGroupMsg().GetGroupUin(),
				Msg:    &proto.MsgReq_PicMsg{PicMsg: &proto.Files{File: []*proto.File{file.File}}},
				AtUin:  nil,
				BotUin: event.GetCurrentQQ(),
			})
			//S.LogInfo(ctx, *resp.ErrMsg)
			if err != nil {
				S.LogError(ctx, err.Error())
				return nil, err
			}

		}
	}

	return &emptypb.Empty{}, nil
}
func (p meme) Unload(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil

}
func (p meme) Init(ctx context.Context, _ *emptypb.Empty) (*proto.InitReply, error) {
	res, err := S.HttpGet(ctx, os.Getenv("memeUrl")+"/memes/keys", map[string]string{"accept": "application/json"})
	if err != nil {
		return nil, err
	}
	keys := &model.KeysRes{}
	err = keys.UnmarshalJSON(res.Content)
	if err != nil {
		return nil, err
	}
	Keys = map[string]*model.KeyInfoRes{}

	for _, v := range *keys {
		res, err = S.HttpGet(ctx, fmt.Sprintf(os.Getenv("memeUrl")+"/memes/%s/info", v), map[string]string{"accept": "application/json"})
		if err != nil {
			return nil, err
		}
		info := &model.KeyInfoRes{}
		err = info.UnmarshalJSON(res.Content)
		if err != nil {
			return nil, err
		}
		Keys[v] = info
		S.LogInfo(ctx, fmt.Sprintf("载入 %s", v))
	}

	return &proto.InitReply{
		Ok:      true,
		Message: "Success",
	}, nil
}

func main() {
	proto.RegisterEvent(meme{})
}
