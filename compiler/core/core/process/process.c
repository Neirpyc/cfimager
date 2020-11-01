//
// Created by cyprien on 25/08/2020.
//

#include "process.h"

void work_list_start(work_list* work_lst)
{
	pthread_t t;

	pthread_create(&t, NULL, &work_list_do, work_lst);
}


void set_canvas(unsigned char* ptr, unsigned int w, unsigned int h)
{
	MAIN_THREAD_EM_ASM({
		let data = Module.HEAPU8.slice($0, $0 + $1 * $2 * 4);
		let canvas = Module['canvas'];
		canvas.width = $1;
		canvas.height = $2;
		let context = canvas.getContext('2d');
		let imageData = context.getImageData(0, 0, $1, $2);
		imageData.data.set(data);
		context.putImageData(imageData, 0, 0);
	}, ptr, w, h);
}
