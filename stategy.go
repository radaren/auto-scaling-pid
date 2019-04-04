package main

import (
	"fmt"
	"math"
)
// baseline: 测试阈值伸缩
func boundStategy (scaleInBound, scaleOutBound float64, 
	initialReplica, minReplica, maxReplica, defaultNs int32) func(int32, float64)(int32, bool){
	if scaleInBound >= scaleOutBound{
		fmt.Println("Error : ", scaleInBound >= scaleOutBound)
	}
	currentReplica := initialReplica

	return func(realPod int32, currentLoad float64) (int32, bool){
		oldReplica := currentReplica
		if currentLoad < scaleInBound {
			if realPod - defaultNs >= minReplica{
				currentReplica = realPod - defaultNs
			}else{
				currentReplica = minReplica
			}
		}else if currentLoad > scaleOutBound {
			if realPod + defaultNs <= maxReplica{
				currentReplica = realPod + defaultNs
			}else{
				currentReplica = maxReplica
			}
		}
		return currentReplica, oldReplica != currentReplica
	}
} 

// baseline: 测试阈值伸缩
func dynamicBoundStategy (scaleInBound, scaleOutBound float64, 
	initialReplica, minReplica, maxReplica int32) func(int32, float64)(int32, bool){
	if scaleInBound >= scaleOutBound{
		fmt.Println("Error : ", scaleInBound >= scaleOutBound)
	}
	currentReplica := initialReplica

	return func(realPod int32, currentLoad float64) (int32, bool){
		oldReplica := currentReplica
		if currentLoad < scaleInBound{
			Nsl := (scaleInBound  - currentLoad) * float64(realPod) / scaleInBound
			Nsh := (scaleOutBound - currentLoad) * float64(realPod) / scaleOutBound
			exceptNode := int32(math.Floor((Nsl + Nsh) / 2 + 0.5))
			if realPod - exceptNode <= minReplica{
				currentReplica = minReplica
			}else{
				currentReplica = realPod - exceptNode
			}
		}else if currentLoad > scaleOutBound{
			Nsl := (currentLoad - scaleOutBound) * float64(realPod) / scaleOutBound
			Nsh := (currentLoad - scaleInBound) * float64(realPod) / scaleInBound
			exceptNode := int32(math.Floor((Nsl + Nsh) / 2 + 0.5))
			if realPod + exceptNode <= maxReplica{
				currentReplica = realPod + exceptNode
			}else{
				currentReplica = maxReplica
			}
		}
		return currentReplica, oldReplica != currentReplica
	}
} 
// P: u_t = u_{t-1} + K_p * e(t)
func PStategy(scaleInBound, scaleOutBound float64, 
	initialReplica, minReplica, maxReplica int32,
	Kp float64) func(int32, float64)(int32, bool){
		currentReplica := initialReplica
		desiredReplica := float64(currentReplica)

		return func(realPod int32, currentLoad float64)(int32, bool){
			et := float64(0)
			oldReplica := currentReplica
			if currentLoad < scaleInBound{
				et = currentLoad - scaleInBound
			}else if currentLoad > scaleOutBound{
				et = currentLoad - scaleOutBound
			}
			desiredReplica = float64(realPod) + et * Kp
			if dR := int32(math.Floor(desiredReplica + 0.5)); currentReplica != dR {
				if dR < minReplica{
					currentReplica = minReplica
				}else if dR > maxReplica{
					currentReplica = maxReplica
				}else{
					currentReplica = dR
				}
			}
			return currentReplica, oldReplica != currentReplica
		}
}
// PI:
func PIStategy(scaleInBound, scaleOutBound float64, 
	initialReplica, minReplica, maxReplica int32,
	Kp, Ki, intergalTime float64) func(int32, float64)(int32, bool){
		currentReplica := initialReplica
		desiredReplica := float64(currentReplica)
		errorSum       := float64(0)

		return func(realPod int32, currentLoad float64)(int32, bool){
			et := float64(0)
			oldReplica := currentReplica
			if currentReplica > 1 && currentLoad < scaleInBound{
				et = currentLoad - scaleInBound
			}else if currentLoad > scaleOutBound{
				et = currentLoad - scaleOutBound
			}
			// if currentReplica == 1 && currentLoad < scaleOutBound{
			// 	et = 0
			// }else{
			// 	et = currentLoad - (scaleInBound + scaleOutBound)/2
			// }
			errorSum += et * intergalTime
			desiredReplica = float64(realPod) + et * Kp + errorSum * Ki

			fmt.Println("desiredReplica: ", desiredReplica, ", calcu replica: ", int32(math.Floor(desiredReplica + 0.5)))
			if dR := int32(math.Floor(desiredReplica + 0.5)); currentReplica != dR {
				if dR < minReplica{
					currentReplica = minReplica
				}else if dR > maxReplica{
					currentReplica = maxReplica
				}else{
					currentReplica = dR
				}
			}
			return currentReplica, oldReplica != currentReplica
		}
}
// PD:
func PDStategy(scaleInBound, scaleOutBound float64, 
	initialReplica, minReplica, maxReplica int32,
	Kp, Kd float64) func(int32, float64)(int32, bool){
		currentReplica := initialReplica
		desiredReplica := float64(currentReplica)
		errorlast      := float64(0)
		return func(realPod int32, currentLoad float64)(int32, bool){
			et := float64(0)
			oldReplica := currentReplica
			if currentReplica > 1 && currentLoad < scaleInBound{
				et = currentLoad - scaleInBound
			}else if currentLoad > scaleOutBound{
				et = currentLoad - scaleOutBound
			}
			desiredReplica = float64(realPod) + et * Kp + errorlast * Kd
			errorlast = et
			fmt.Println("desiredReplica: ", desiredReplica, ", calcu replica: ", int32(math.Floor(desiredReplica + 0.5)))
			if dR := int32(math.Floor(desiredReplica + 0.5)); currentReplica != dR {
				if dR < minReplica{
					currentReplica = minReplica
				}else if dR > maxReplica{
					currentReplica = maxReplica
				}else{
					currentReplica = dR
				}
			}
			return currentReplica, oldReplica != currentReplica
		}
}
// PID:
func PIDStategy(scaleInBound, scaleOutBound float64, 
	initialReplica, minReplica, maxReplica int32,
	Kp, Ki, Kd, intergalTime float64) func(int32, float64)(int32, bool){
		currentReplica := initialReplica
		desiredReplica := float64(currentReplica)
		errorlast      := float64(0)
		errorSum       := float64(0)
		return func(realPod int32, currentLoad float64)(int32, bool){
			et := float64(0)
			oldReplica := currentReplica
			if currentReplica > 1 && currentLoad < scaleInBound{
				et = currentLoad - scaleInBound
			}else if currentLoad > scaleOutBound{
				et = currentLoad - scaleOutBound
			}
			errorSum += et * intergalTime
			desiredReplica = float64(realPod) + et * Kp + errorSum * Ki + errorlast * Kd
			errorlast = et
			fmt.Println("desiredReplica: ", desiredReplica, ", calcu replica: ", int32(math.Floor(desiredReplica + 0.5)))
			if dR := int32(math.Floor(desiredReplica + 0.5)); currentReplica != dR {
				if dR < minReplica{
					currentReplica = minReplica
				}else if dR > maxReplica{
					currentReplica = maxReplica
				}else{
					currentReplica = dR
				}
			}
			return currentReplica, oldReplica != currentReplica
		}
}