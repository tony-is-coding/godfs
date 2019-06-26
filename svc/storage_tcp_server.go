package svc

import (
	"encoding/json"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/uuid"
	"github.com/logrusorgru/aurora"
	"io"
	"net"
	"strings"
	"time"
)

func StartStorageTcpServer() {
	listener, err := net.Listen("tcp", common.Config.BindAddress+":"+convert.IntToStr(common.Config.Port))
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Millisecond * 50)
	logger.Info("  tcp server starting on port ", common.Config.Port)
	logger.Info(aurora.BrightGreen(":::server started:::"))
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting new connection: ", err)
			continue
		}
		logger.Debug("accept a new connection")
		gox.Try(func() {
			go clientConnHandler(conn)
		}, func(i interface{}) {
			logger.Error("connection error:", err)
		})
	}
}

func clientConnHandler(conn net.Conn) {
	pip := &gpip.Pip{
		Conn: conn,
	}
	defer pip.Close()
	validated := false
	for {
		err := pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
			header := _header.(*common.Header)
			bs, _ := json.Marshal(header)
			logger.Debug("server got message:", string(bs))
			if header.Operation == common.OPERATION_CONNECT {
				h, b, l, err := authenticationHandler(header)
				if err != nil {
					return err
				}
				if h.Result != common.SUCCESS {
					pip.Send(h, b, l)
					return errors.New("unauthorized connection, force disconnection by server")
				} else {
					validated = true
					return pip.Send(h, b, l)
				}
			} else if header.Operation == common.OPERATION_UPLOAD {
				h, b, l, err := uploadFileHandler(bodyReader, bodyLength, validated)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			}
			return pip.Send(&common.Header{
				Result:     common.SUCCESS,
				Msg:        "",
				Attributes: map[string]interface{}{"Name": "李四"},
			}, nil, 0)
		})
		if err != nil {
			logger.Error("error receive data:", err)
			pip.Close()
			break
		}
	}
}

func authenticationHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, 0, nil
	}
	secret := header.Attributes["secret"]
	if secret != common.Config.Secret {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, 0, nil
	}
	return &common.Header{
		Result: common.SUCCESS,
		Msg:    "authentication success",
	}, nil, 0, nil
}

func uploadFileHandler(bodyReader io.Reader, bodyLength int64, authorized bool) (*common.Header, io.Reader, int64, error) {
	if !authorized {
		return nil, nil, 0, errors.New("unauthorized connection")
	}
	buffer := make([]byte, common.BUFFER_SIZE)
	var realRead int64 = 0
	crcH := util.CreateCrc32Hash()
	md5H := util.CreateMd5Hash()
	tmpFileName := common.Config.TmpDir + "/" + uuid.UUID()
	out, err := file.CreateFile(tmpFileName)
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() {
		defer out.Close()
		file.Delete(tmpFileName)
	}()
	for true {
		n, err := bodyReader.Read(buffer)
		realRead += int64(n)
		if err != nil && err != io.EOF {
			return nil, nil, 0, err
		}
		if n > 0 {
			_, err := crcH.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
			_, err = md5H.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
			_, err = out.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
		} else {
			out.Close()
			if bodyLength != realRead {
				return nil, nil, 0, errors.New("mismatch body length")
			}
			crc32String := util.GetCrc32HashString(crcH)
			md5String := util.GetMd5HashString(md5H)

			targetDir := strings.ToUpper(strings.Join([]string{crc32String[len(crc32String)-4 : len(crc32String)-2], "/",
				crc32String[len(crc32String)-2:]}, ""))
			// 文件放在crc结尾的目录，防止目恶意伪造md5文件进行覆盖
			// 避免了暴露文件md5可能出现的风险：保证了在md5相等但是文件不同情况下文件出现的覆盖情况。
			// 此时要求文件的交流必须携带完整的参数
			targetLoc := common.Config.DataDir + "/" + targetDir
			targetFile := common.Config.DataDir + "/" + targetDir + "/" + md5String
			finalFileId := common.Config.Group + "/" + targetDir + "/" + md5String
			if !file.Exists(targetLoc) {
				if err := file.CreateDirs(targetLoc); err != nil {
					return nil, nil, 0, err
				}
			}
			if file.Exists(targetFile) {
				return &common.Header{
					Result:     common.SUCCESS,
					Attributes: map[string]interface{}{"fid": finalFileId},
				}, nil, 0, nil
			}
			if err := file.MoveFile(tmpFileName, targetFile); err != nil {
				return nil, nil, 0, err
			}
			return &common.Header{
				Result:     common.SUCCESS,
				Attributes: map[string]interface{}{"fid": finalFileId},
			}, nil, 0, nil
		}
	}

	return &common.Header{}, nil, 0, nil
}
