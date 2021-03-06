/*
 * autoscaling
 *
 * <br/>https://ncloud.apigw.ntruss.com/autoscaling/v2
 *
 * API version: 2018-08-07T06:47:31Z
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package autoscaling

type SetDesiredCapacityRequest struct {

	// 오토스케일링그룹명
AutoScalingGroupName *string `json:"autoScalingGroupName"`

	// 기대용량치
DesiredCapacity *int32 `json:"desiredCapacity"`
}
