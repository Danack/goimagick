// Copyright 2013 Herbert G. Fischer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imagick

/*
#include <magick/MagickCore.h>
*/
import "C"

//import "runtime"
import "unsafe"

// This struct represents the KernelInfo C API of ImageMagick
type KernelInfo struct {
	info *C.KernelInfo
}

func newKernelInfo(cki *C.KernelInfo) *KernelInfo {
	ki := &KernelInfo{info: cki}
	//TODO - understand and implement this.
	// runtime.SetFinalizer(ki, Destroy)
	//ki.IncreaseCount()
	return ki
}

// Convert the current KernelInfo to an 2d-array of values. The values are either
// a float64 if the element is used, or NaN if the element is not used by the kernel
func (kernel_info *KernelInfo) ToArray() [][]float64 {
	count := 0
	values := [][]float64{}

	for y := C.size_t(0); y < kernel_info.info.height; y++ {
		row_values := make([]float64, kernel_info.info.width)
		for x := C.size_t(0); x < kernel_info.info.width; x++ {
			p2 := (*[1 << 10]C.double)(unsafe.Pointer(kernel_info.info.values))
			row_values[x] = float64(p2[count])
			count++
		}
		values = append(values, row_values)
	}

	return values
}

// Create a kernel from a builtin in kernel. See http://www.imagemagick.org/Usage/morphology/#kernel
// for examples. Currently the 'rotation' symbols are not supported. Example:
// kernel_info := imagick.AcquireKernelBuiltIn(imagick.KERNEL_RING, "2,1")
func AcquireKernelBuiltIn(kernelType KernelInfoType, kernelString string) *KernelInfo {
	gi := C.GeometryInfo{}
	cKernelString := C.CString(kernelString)
	defer C.free(unsafe.Pointer(cKernelString))
	result := C.ParseGeometry(cKernelString, &gi)
	var geometryFlags int = int(result)
	FiddleWithGeometryInfo(kernelType, geometryFlags, &gi)
	kernelInfo := C.AcquireKernelBuiltIn(C.KernelInfoType(kernelType), &gi)

	return newKernelInfo(kernelInfo)
}

// ScaleKernelInfo() scales the given kernel list by the given amount, with or without
// normalization of the sum of the kernel values (as per given flags). The exact behaviour
// of this function depends on the normalization type being used please see
// http://www.imagemagick.org/api/morphology.php#ScaleKernelInfo for details.
// Flag should be one of:
// imagick.KERNEL_NORMALIZE_NONE
// imagick.KERNEL_NORMALIZE_VALUE
// imagick.KERNEL_NORMALIZE_CORRELATE
// imagick.KERNEL_NORMALIZE_PERCENT
func (kernel_info *KernelInfo) Scale(scale float64, normalize_type KernelNormalizeType) {
	C.ScaleKernelInfo(kernel_info.info, C.double(scale), C.GeometryFlags(normalize_type))
}

// This does .....stuff. Basically some tidy up of the kernel is required apparently.
func FiddleWithGeometryInfo(kernelType KernelInfoType, geometryFlags int, geometryInfo *C.GeometryInfo) {
	/* special handling of missing values in input string */
	switch kernelType {
	/* Shape Kernel Defaults */
	case KERNEL_UNITY:
		if (geometryFlags & WIDTHVALUE) == 0 {
			geometryInfo.rho = 1.0 /* Default scale = 1.0, zero is valid */
		}
	case KERNEL_SQUARE:
	case KERNEL_DIAMOND:
	case KERNEL_OCTAGON:
	case KERNEL_DISK:
	case KERNEL_PLUS:
	case KERNEL_CROSS:
		if (geometryFlags & HEIGHTVALUE) == 0 {
			geometryInfo.sigma = 1.0 /* Default scale = 1.0, zero is valid */
		}
	case KERNEL_RING:
		if (geometryFlags & XVALUE) == 0 {
			geometryInfo.xi = 1.0 /* Default scale = 1.0, zero is valid */
		}
	case KERNEL_RECTANGLE:
		/* Rectangle - set size defaults */
		if (geometryFlags & WIDTHVALUE) == 0 { /* if no width then */
			geometryInfo.rho = geometryInfo.sigma /* then  width = height */
		}
		if geometryInfo.rho < 1.0 { /* if width too small */
			geometryInfo.rho = 3 /* then  width = 3 */
		}
		if geometryInfo.sigma < 1.0 { /* if height too small */
			geometryInfo.sigma = geometryInfo.rho /* then  height = width */
		}
		if ((geometryFlags & XVALUE) == 0) {    /* center offset if not defined */
			geometryInfo.xi = C.double((int(geometryInfo.rho)-1)/2);
		}
		if ((geometryFlags & YVALUE) == 0) {
			geometryInfo.psi = C.double((int(geometryInfo.sigma)-1)/2);
		}
	/* Distance Kernel Defaults */
	case KERNEL_CHEBYSHEV:
	case KERNEL_MANHATTAN:
	case KERNEL_OCTAGONAL:
	case KERNEL_EUCLIDEAN:
		if (geometryFlags & HEIGHTVALUE) == 0 { /* no distance scale */
			geometryInfo.sigma = 100.0 /* default distance scaling */
		} else if (geometryFlags & ASPECTVALUE) != 0 {     /* '!' flag */
			geometryInfo.sigma = QUANTUM_RANGE / (geometryInfo.sigma+1) /* maximum pixel distance */
		} else if (geometryFlags & PERCENTVALUE) != 0 {    /* '%' flag */
			geometryInfo.sigma *= QUANTUM_RANGE / 100.0         /* percentage of color range */
		}
	default:
	}
}
