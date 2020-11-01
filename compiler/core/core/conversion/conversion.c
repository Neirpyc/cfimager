//
// Created by cyprien on 26/08/2020.
//

#include "conversion.h"
complex double strtoc(const char* str, char** endptr)
{
	char* ptr0;
	double real, imag, * second_fill;

	second_fill = &imag;
	errno = 0;
	real = strtod(str, &ptr0);

	if (real == 0 && str == ptr0)
		errno = 1;
	if (errno != 0)
		return 0;
	while (isspace(*ptr0))
	{ ptr0++; }
	if (*ptr0 == 'i')
	{
		imag = real;
		second_fill = &real;
		ptr0++;
	}
	*second_fill = strtod(ptr0, &ptr0);
	if (*second_fill == 0 && str == ptr0)
		errno = 1;
	if (errno != 0)
		return 0;
	if (second_fill == &imag && *ptr0 != 'i' || second_fill == &real && *ptr0 == 'i')
	{
		errno = EINVAL;
		return 0;
	}

	if (endptr != NULL)
	{
		*endptr = ++ptr0;
	}

	return real + imag * I;
}

range strto_range(char* str)
{
	range result;

	if (*str == '\'' || *str == '\"')
	{ str++; }

	while (isspace(*str))
	{ str++; }

	if (*(str++) != '{')
	{
		errno = EINVAL;
		return result;
	}

	errno = 0;
	result.top_left = strtoc(str, &str);
	if (errno != 0)
	{
		return result;
	}

	while (isspace(*str))
	{ str++; }

	if (*(str++) != ';')
	{
		errno = EINVAL;
		return result;
	}

	result.bot_right = strtoc(str, &str);

	return result;
}
dimension strto_dimension(char* str)
{
	dimension result;
	
	if (sscanf(str, "%ux%u", &result.width, &result.height) != 2)
	{
		errno = EINVAL;
	}

	return result;
}
