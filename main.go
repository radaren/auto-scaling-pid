// 基于php-apache的auto-scaling框架
// framework for auto-scaling based on php-apache

package main

import(
	"os"
	"fmt"
	
	"time"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/csv"
	"k8s.io/client-go/util/retry"
	"k8s.io/apimachinery/pkg/api/resource"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


func main(){
	file, err := os.Create("p_r.csv")
	if err != nil{
		panic(err)
	}
    defer file.Close()

	writer := csv.NewWriter(file)
	writer.Write([]string{"Ticker", "QPS", "cpuCost", "realPod", "exceptPod","cost"})		
    defer writer.Flush()


	// 初始化环境变量
	clientset := initClientSet()
	deploymentClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	// 初始化deployment
	_, err = deploymentClient.Create(generateDeployment())
	if err != nil{
		panic(err)
	}
	fmt.Println("please wait about 70s for initial replica start")
	time.Sleep(70 * time.Second)

	// 设置参数
	scaleInBound   		:= 20.0
	scaleOutBound  		:= 40.0
	initialReplica 		:= int32(2)
	minReplica     		:= int32(1)
	maxReplica     		:= int32(128)
	// defaultNs           := int32(2)
	Kp                  := 0.07
	// Ki                  := 0.002
	// Kd                  := 0.0001
	periodTime := int64(10)
	periodSum  := 0.0

	// 策略模型
	// BoundHPA  := boundStategy(scaleInBound, scaleOutBound, initialReplica, minReplica, maxReplica, defaultNs)
	// dBoundHPA := dynamicBoundStategy(scaleInBound, scaleOutBound, initialReplica, minReplica, maxReplica)
	PHPA      := PStategy(scaleInBound, scaleOutBound, initialReplica, minReplica, maxReplica, Kp)
	// PIHPA     := PIStategy(scaleInBound, scaleOutBound, initialReplica, minReplica, maxReplica, Kp, Ki, float64(periodTime))
	// PDHPA     := PDStategy(scaleInBound, scaleOutBound, initialReplica, minReplica, maxReplica, Kp, Kd)
	// PIDHPA       := PIDStategy(scaleInBound, scaleOutBound, initialReplica, minReplica, maxReplica, Kp, Ki, Kd, float64(periodTime))
	
	// 请求模型
	RectReq  := rectGenerator(2, 8)
	// SimReq   := sinGenerator(5.0, 0.1)
	// expReq   := expGenerator(0.8, 0.001)
	// logReq   := logGenerator(2.0, 2.0)
	var counterTime = int64(0)
	ticker := time.NewTicker(time.Second) // 每秒同步一次
	expCounter := initialReplica

	for range ticker.C {
		var pods PodMetricsList
		data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw()
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(data, &pods)
		if err != nil {
			fmt.Println(err)
		}
		cpuCounter := int32(0)
		cpuCost := 0.0
		for _, m := range pods.Items {
			if strings.HasPrefix(m.Metadata.Name, "php-apache"){
				// fmt.Println(m.Metadata.Name, m.Metadata.Namespace, m.Timestamp.String())
				for _, c := range m.Containers{
					//fmt.Println(c.Usage.CPU, c.Usage.Memory)
					cCPU, err := strconv.ParseFloat(strings.Trim(c.Usage.CPU, "n"), 64)
					if err == nil{
						cpuCounter = cpuCounter + 1
						cpuCost = cpuCost + cCPU / 1000000.0	
					}
				}		
			}
		}
		cpuAvg := 0.0
		if cpuCounter > 0{
			cpuAvg = cpuCost / float64(cpuCounter)
		}

		reqNum := RectReq(counterTime)
		if periodSum > 0 && counterTime % periodTime == 0{
			if changeVal, shouldChange := PHPA(cpuCounter, periodSum / float64(periodTime)); shouldChange{
				fmt.Println("Info: change replica to", changeVal)
				retry.RetryOnConflict(retry.DefaultRetry, func() error{
					obj, err := deploymentClient.Get("php-apache", metav1.GetOptions{})
					if err != nil{
						return err
					}
					obj.Spec.Replicas = int32Ptr(changeVal)
					_, err = deploymentClient.Update(obj)
					return err
				})
				expCounter = changeVal
			}
			periodSum = 0	
		}else{
			periodSum += cpuAvg
		}
		cost := float64(0)
		if cpuCounter > 1 && cpuAvg < scaleInBound{
			cost = (scaleInBound - cpuAvg) * float64(cpuCounter)
		}
		if cpuAvg > scaleOutBound{
			cost = (cpuAvg - scaleOutBound) * float64(cpuCounter)
		}
		writer.Write([]string{strconv.FormatInt(counterTime, 10), 
			strconv.FormatInt(reqNum,10), fmt.Sprintf("%f", cpuAvg), 
			strconv.FormatInt(int64(cpuCounter),10),strconv.FormatInt(int64(expCounter),10),
			fmt.Sprintf("%f", cost)})
		
		fmt.Println("Ticker:", counterTime, 
			",\tvalid Replica:", cpuCounter,
			",\tavg cpuCost:", cpuAvg,
			",\tgen Request:", reqNum,
			",\tcost:", cost)
		counterTime = counterTime + 1
		if counterTime > 1000{
			break
		}
	}
	
	fmt.Println("Press Enter to Delete deployment")
	prompt()
	// 删除 deployment
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentClient.Delete("php-apache", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
}


func generateDeployment()*appsv1.Deployment{
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "php-apache",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app" : "php-apache",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "php-apache",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "php-apache",
							Image: "pilchard/hpa-example",
							ImagePullPolicy: apiv1.PullNever, // use Local image
							Ports: []apiv1.ContainerPort{
								{
									Name: "http",
									ContainerPort: 80,
									Protocol: apiv1.ProtocolTCP,	
								},
							},
							Resources: apiv1.ResourceRequirements{
								// Requests: apiv1.ResourceList{
								// 	"cpu":    resource.MustParse("100m"),
								// },
								Limits: apiv1.ResourceList{
									"cpu":    resource.MustParse("100m"),
									// "memory": resource.MustParse("500m"),
								},
							},	
						},
					},
					RestartPolicy: apiv1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: int64Ptr(30),
				},
			},
		},
	}
}