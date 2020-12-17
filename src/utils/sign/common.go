package sign

import (
	
)

/**
 * 获得一个secp256k1公密钥对
 */
 func GenKeys() interface{}{

	return nil
}

/**
 * 通过一个合法密钥字符串获得公钥
 * @param privateKey string secp256k1密钥
 */
func GetPublicKey (privateKey string) string{

	return ""
}

/**
 * 利用私钥加密一个字符串的函数。
 * @param msg string 需要加密的字符串
 * @param privateKey string 密钥
 */
func Sign (msg string, privateKey string) string {

	return ""
}

/**
 * 校验一个签名是否使用该pk对该msg签注的
 * @param signature string 签名字符串
 * @param msg string 被签名的字符串
 * @param publicKey string 公钥
 */
func Verify (signature string, msg string, publicKey string) bool{

	return true
}

/**
 * 从签名后的信息中提取中签名的公钥
 * @param signature string 签名字符串
 * @param rcid uint64 签名回复编号
 * @param msg string 被签注的消息
 */
func Recover (signature string, rcid uint64, msg string) string {

	return ""
}