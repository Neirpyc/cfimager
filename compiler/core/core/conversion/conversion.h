//
// Created by cyprien on 26/08/2020.
//

#ifndef _CONVERSION_H_
#define _CONVERSION_H_

#include <complex.h>
#include <ctype.h>
#include <stdlib.h>
#include <errno.h>
#include "../../set_render_settings.h"

complex double strtoc(const char* str, char** endptr);

range strto_range(char *str);

dimension strto_dimension(char *str);

#endif //_CONVERSION_H_
