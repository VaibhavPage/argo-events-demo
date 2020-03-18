import cv2
import numpy

src = cv2.imread('/home/argo-logo.png')
 
src = cv2.GaussianBlur(src,(filter_a,filter_b),cv2.BORDER_DEFAULT)

s_vs_p = 0.7
amount = 0.010
out = numpy.copy(src)
# Salt mode
num_salt = numpy.ceil(amount * src.size * s_vs_p)
coords = [numpy.random.randint(0, i - 1, int(num_salt))
      for i in src.shape]
out[coords] = 1

# Pepper mode
num_pepper = numpy.ceil(amount* src.size * (1. - s_vs_p))
coords = [numpy.random.randint(0, i - 1, int(num_pepper))
      for i in src.shape]
out[coords] = 0

dst = out

cv2.imwrite('/home/vaibhav/Documents/blur-argo-logo.png', dst)
