#include "component.h"
//some tools

/*********************
 *add to dispatcher::swap2
 *********************/
void dispatcher::tool_pcbStatistics(){
	FILE *pnt;
	int i;
	int sum=0;
	static int flag=0;
	
	pnt=fopen("pcb_statistics.txt","a");

	if(flag==0)
		fprintf(pnt,"---process control block statistics---\n\nused now  ready  wait  total(add run)\n");
	
	flag=1;
	for(i=0; i<10; i++)	//total wait process
		sum+=waitb[i].len;

	fprintf(pnt," %-10d%-8d%-8d%-d\n",usedPcbNumber,readyb.len,sum,readyb.len+sum+1);

	//if(readyb.len==0){
	//	fprintf(pnt,"---END---\n\n");
		//fclose(pnt);
	//}
	//havent close file for some error
}

int Load(procptr pptr, char *s){
	FILE *fo;		
	char tmps[50]; //用了保存新的文件名
	char *p;
	char ext[10]=".co"; //字符码文件的扩展名为co

	strcpy(tmps,s);
	p=tmps;

	while(*p!='.') p++;	//将字符码文件的扩展名改为xo
	*p='\0';
	strcat(tmps,ext);
	
	fo=fopen(tmps, "r");

	int cw;
	int i=0;
	do{
		fscanf(fo,"%d",&cw);
		pptr->MCode.write(i,cw);
		i++;
	}while(cw!=900000);

	int len;
	fscanf(fo,"%d",&len);

	for(i=0; i<len; i++){
		fscanf(fo,"%d",&cw);
		pptr->MData.write(i,cw);
	}
	
	
	return 0;
}