/*
 * server
 *
 * <br/>https://ncloud.apigw.ntruss.com/server/v2
 *
 * API version: 2018-08-31T04:58:16Z
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package server

type ChangeNasVolumeSizeRequest struct {

	// NAS볼륨인스턴스번호
NasVolumeInstanceNo *string `json:"nasVolumeInstanceNo"`

	// NAS볼륨사이즈
VolumeSize *int32 `json:"volumeSize"`
}
