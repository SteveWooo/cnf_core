package sign

// 参考文档：https://godoc.org/github.com/btcsuite/btcd/btcec#ParseSignature

import (
	"crypto/sha256"
	"encoding/hex"

	btcec "github.com/cnf_core/pkg/btcec"
	error "github.com/cnf_core/src/utils/error"
)

/**
 * 获得一个secp256k1公密钥对
 */
func GenKeys() interface{} {
	privateKey, _ := btcec.NewPrivateKey(btcec.S256())
	publicKey := privateKey.PubKey()

	keys := map[string]string{
		"publicKey":  hex.EncodeToString(publicKey.SerializeUncompressed()),
		"privateKey": hex.EncodeToString(privateKey.Serialize()),
	}

	return keys
}

/**
 * 通过一个合法密钥字符串获得公钥
 * @param privateKey string secp256k1密钥
 */
func GetPublicKey(privateKey string) string {
	privateKeyByte, _ := hex.DecodeString(privateKey)
	_, publicKey := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyByte)

	return hex.EncodeToString(publicKey.SerializeUncompressed())
}

/**
 * 利用私钥加密一个字符串的函数。
 * @param msg string 需要加密的字符串
 * @param privateKey string 密钥
 */
func Sign(msg string, privateKeyStr string) (string, interface{}) {
	if len(privateKeyStr) != 64 {
		return "", error.New(map[string]interface{}{
			"message": "不合法私钥",
		})
	}

	if len(msg) != 64 {
		return "", error.New(map[string]interface{}{
			"message": "签名内容不合法",
		})
	}

	privateKeyByte, _ := hex.DecodeString(privateKeyStr)
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyByte)

	msgByte, _ := hex.DecodeString(msg)

	// 使用可以从签名中恢复出公钥的方式签名。
	signature, signErr := btcec.SignCompact(btcec.S256(), privateKey, msgByte, false)
	if signErr != nil {
		return "", error.New(map[string]interface{}{
			"message":   "签名失败",
			"originErr": signErr,
		})
	}
	return hex.EncodeToString(signature), nil
}

/**
 * 🚮废弃函数
 * 校验一个签名是否使用该pk对该msg签注的
 * @param signature string 签名字符串
 * @param msg string 被签名的字符串，一般加密的都是摘要
 * @param publicKey string 公钥
 */
func Verify(signatureStr string, msg string, publicKeyStr string) bool {
	signBytes, _ := hex.DecodeString(signatureStr)
	signature, _ := btcec.ParseSignature(signBytes, btcec.S256())

	msgByte, _ := hex.DecodeString(msg)

	publicKeyByte, _ := hex.DecodeString(publicKeyStr)
	publicKey, _ := btcec.ParsePubKey(publicKeyByte, btcec.S256())

	verifyed := signature.Verify(msgByte, publicKey)

	return verifyed
}

/**
 * 从签名后的信息中提取中签名的公钥
 * @param signature string 签名字符串
 * @param rcid uint64 签名回复编号
 * @param msg string 被签注的消息
 */
func Recover(signatureStr string, msg string, recid uint64) (string, interface{}) {
	if len(msg) != 64 {
		return "", error.New(map[string]interface{}{
			"message": "签名内容不合法",
		})
	}

	msgByte, _ := hex.DecodeString(msg)

	if recid == 1 {
		signatureStr = "1c" + signatureStr
	}

	if recid == 0 {
		signatureStr = "1b" + signatureStr
	}

	signBytes, _ := hex.DecodeString(signatureStr)

	publicKey, _, recoverErr := btcec.RecoverCompact(btcec.S256(), signBytes, msgByte)
	if recoverErr != nil {
		return "", error.New(map[string]interface{}{
			"message":   "公钥恢复失败",
			"originErr": recoverErr,
		})
	}

	return hex.EncodeToString(publicKey.SerializeUncompressed()), nil
}

/**
 * 对字符串进行sha256哈希
 */
func Hash(msg string) string {
	hashByte := sha256.Sum256([]byte(msg))
	return hex.EncodeToString(hashByte[:])
}
