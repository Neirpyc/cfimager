//
// Created by cyprien on 25/08/2020.
//

#ifndef _PROCESS_H_
#define _PROCESS_H_

#include <complex.h>
#include <math.h>
#include "../image/image.h"
#include "../../set_render_settings.h"

void set_canvas(unsigned char* ptr, unsigned int w, unsigned int h);
void work_list_start(work_list *work_lst);

#include "async.h"

#endif //_PROCESS_H_
