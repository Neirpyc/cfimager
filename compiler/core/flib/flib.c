#include <complex.h>

#define ITERATIONS 10

extern double complex f(double complex c) {
	double complex k;
	double complex sum;
	unsigned int i;

	sum = 1;
	k = 1 / (1 - cpow(2, 1 - c));
	for (i = 2; i < ITERATIONS / 2; i += 2) {
		sum -= 1 / cpow(i, c);
		sum += 1 / cpow(i + 1, c);
	}
	return k * sum;
}
