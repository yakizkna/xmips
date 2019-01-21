#include "component.h"

int assembler::trans(char *op, int line){	//translate aux to digit
	if(!strcmp(op, "MOV") || !strcmp(op, "mov"))     return 900152;
	//if(!strcmp(op, "IMMOV") || !strcmp(op, "immov")) return 90112;	//IMMOV, M, and ASN image to same code
	//if(!strcmp(op, "M") || !strcmp(op, "m"))         return 90112;
	//if(!strcmp(op, "ASN") || !strcmp(op, "asn"))     return 90112; 
	if(!strcmp(op, "ADD") || !strcmp(op, "add"))     return 900272;
	if(!strcmp(op, "SUB") || !strcmp(op, "sub"))     return 901072;
	if(!strcmp(op, "MUL") || !strcmp(op, "mul"))     return 902072;
	if(!strcmp(op, "DIV") || !strcmp(op, "div"))     return 903072;
	if(!strcmp(op, "MOD") || !strcmp(op, "mod"))     return 904072;
	if(!strcmp(op, "AND") || !strcmp(op, "and"))     return 905072;
	if(!strcmp(op, "OR") || !strcmp(op, "or"))       return 906072;
	if(!strcmp(op, "XOR") || !strcmp(op, "xor"))     return 907072;
	//if(!strcmp(op, "IMADD") || !strcmp(op, "imadd")) return 901272;
	//if(!strcmp(op, "A") || !strcmp(op, "a"))         return 901272;
	if(!strcmp(op, "CMP") || !strcmp(op, "cmp"))     return 900332;
	//if(!strcmp(op, "IMCMP") || !strcmp(op, "imcmp")) return 90132;
	//if(!strcmp(op, "C") || !strcmp(op, "c"))         return 90132;
	if(!strcmp(op, "JA") || !strcmp(op, "ja"))       return 900421;
	if(!strcmp(op, "JB") || !strcmp(op, "jb"))       return 900521;
	if(!strcmp(op, "JMP") || !strcmp(op, "jmp"))     return 900621;
	if(!strcmp(op, "JE") || !strcmp(op, "je"))       return 900721;
	if(!strcmp(op, "JNE") || !strcmp(op, "jne"))     return 900821;
	if(!strcmp(op, "INC") || !strcmp(op, "inc"))     return 900961;
	if(!strcmp(op, "NEG") || !strcmp(op, "neg"))     return 901961;
	if(!strcmp(op, "NOT") || !strcmp(op, "not"))     return 902961;
	//if(!strcmp(op, "IMSUB") || !strcmp(op, "imsub")) return 90142;
	//if(!strcmp(op, "S") || !strcmp(op, "s"))         return 90142;
	if(!strcmp(op, "LEA") || !strcmp(op, "lea"))     return 902042;	//has better way now
	if(!strcmp(op, "PUSH") || !strcmp(op, "push"))   return 902121;
	if(!strcmp(op, "POP") || !strcmp(op, "pop"))     return 902241;
	if(!strcmp(op, "PUSHA") || !strcmp(op, "pusha")) return 902300;
	if(!strcmp(op, "POPA") || !strcmp(op, "popa"))   return 902400;
	//if(!strcmp(op, "OFST") || !strcmp(op, "ofst"))   return 90302;
	if(!strcmp(op, "LOC") || !strcmp(op, "loc"))     return 910002;
	if(!strcmp(op, "DIM") || !strcmp(op, "dim"))     return 910002; //dim is newer than loc
	if(!strcmp(op, "END") || !strcmp(op, "end"))     return 900000;
	if(!strcmp(op, "INT") || !strcmp(op, "int"))	 return 904000;
	if(!strcmp(op, "CALL") || !strcmp(op, "call"))   return 904121;
	if(!strcmp(op, "RET") || !strcmp(op, "ret"))	 return 904200;
	if(!strcmp(op, "IRET") || !strcmp(op, "iret"))   return 904300;
	if(!strcmp(op, "SYSR") || !strcmp(op, "sysr"))   return 950000;
	if(!strcmp(op, "SYSW") || !strcmp(op, "sysw"))   return 950100;
	if(!strcmp(op, "WAKE") || !strcmp(op, "wake"))   return 950200;
	if(!strcmp(op, "SET") || !strcmp(op, "set"))     return 950300;
	if(!strcmp(op, "CLI") || !strcmp(op, "cli"))     return 960000;
	if(!strcmp(op, "STI") || !strcmp(op, "sti"))     return 960100;
    if(!strcmp(op, "HALT") || !strcmp(op, "halt"))   return 800000;
	if(!strcmp(op, "~"))     return 3; //comment
	if(!strcmp(op, "CMNT") || !strcmp(op, "cmnt"))   return 3; //also comment
	if(!strcmp(op, "$"))     return 0; //filler
	if(!strcmp(op, ":"))	return 920001; //标号语句
	if(!strcmp(op, "PROC") || !strcmp(op, "proc"))   return 920001; //use tag

	//RES=50;
	//if(systemChecker.showLevel(RES, sysLog)){
	//	printf("D%-5d row:%d, opcode:\'%s\' ",ID,line,op);
	//	systemChecker.check(RES, sysLog);
	//}
	return -1;
}



//形式地址生成器
//&表示偏移寻址（只能相对于寄存器1~9），权值为8，并加上寄存器号乘以100。例：&3，对应的值为8+3*100=308
//@表示间接寻址，权值为4
//#表示寄存器寻址，权值为2
//！表示立即数，权值为3
//用一个整型量addressing带回寻址方式
//多重寻址方式可能同时存在，addressing为所有寻址方式权值的和
//例如 &4@#2，包含偏移寻址，间接寻址和寄存器寻址，EA=(R4)+(R2) addressing=8+4*100+4+2=414
//&4#2 是无效的，因为寄存器2只有符号地址，没有实际地址!
//&5@100 EA=(R5)+(100)
//@!32 是无效的，立即数必须直接使用!，如!5。
int assembler::trans2(char *op, varNote *vp, int line,int &addressing){	//ptrflag = 1 means op is ptr var
	int i;
	char *ptmp=op;	//指向操作数地址的指针
	int advisit=0;	//用于计算寻址方式
	int ptrflag; //是否采用了间接寻址方式

	addressing=0;

	if(!strcmp(op,"$")){
		addressing=0;	
		return 0;	//占位符，现在已经可以不用
	}

	if(*ptmp=='&'){	//偏移寻址
		advisit+=8;	//
		if(*(ptmp+1)>='1' && *(ptmp+1)<='9'){	//只能相对寄存器1~9
			advisit+=100*(*(ptmp+1)-'0');	//加上寄存器号乘以100
			ptmp+=2;		//指向操作数字符串中的下一个字符
		}
		else{	//表达式错误，错误地使用了符号&
			RES=53;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
				systemChecker.check(RES, sysLog);
			}
			return -1;
		}
	}

	if(*ptmp=='@'){	//间接寻址
		advisit+=4;
		ptmp++;
		ptrflag=1;	//表示采用了间接寻址方式，
	}
	
	if(*ptmp=='#' || *ptmp=='!'){
		//advisit+=2;!!!
		if(*ptmp=='#'){	//寄存器寻址
			if(advisit % 10==8){	//出现了如&4#2的形式，该形式是无效的，因为寄存器2只有符号地址，没有实际地址! 
				RES=54;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
					systemChecker.check(RES, sysLog);
				}
				return -1;	
			}
			if( *(ptmp+1)>='0' && *(ptmp+1)<='9' ){	//'#'符号后必须为一个代表寄存器号的数字
				int registerNO;
				advisit+=2;
				addressing=advisit;
				
				registerNO=atoi(ptmp+1);
				if(registerNO>14){
					RES=58;
					if(systemChecker.showLevel(RES, sysLog)){
						printf("D%-5d row:%d, register No:%d ",ID,line,registerNO);
						systemChecker.check(RES, sysLog);
					}
					return -1;
				}
				else
					return registerNO;
			}
			else{	//寄存器寻址表达式错误
				RES=55;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
					systemChecker.check(RES, sysLog);
				}
				return -1;
			}
		}
		else{	//改为主存寻址!!! //立即数
			/*
			if((advisit-2)!=0){	//立即数之前出现了其他符合
				RES=56;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
					systemChecker.check(RES, sysLog);
				}
				return -1;	
			}
			*/

			//advisit++;

			//if(*(ptmp+1)=='-' || *(ptmp+1)=='+'|| (*(ptmp+1)>='0' && *(ptmp+1)<='9')){	//立即数的第一个字符必须是数字或正负号
			//	addressing=advisit;
			//	return atoi(ptmp+1);
			//}
			if((*(ptmp+1)>='0' && *(ptmp+1)<='9')){	//立即数的第一个字符必须是数字或正负号
				addressing=advisit;
				return atoi(ptmp+1);
			}


			else{	//'!'符号之后的格式错误
				RES=57;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
					systemChecker.check(RES, sysLog);
				}
				return -1;
			}
		}
	}

	else{	//改为形式地址为主存符号地址或者立即数	//形式地址就是主存地址（包括符号地址）
		i=0;
		while(vp[i].valid==1){	//查找符号地址表
			if(!strcmp(ptmp,vp[i].var)){	//该标示符是已经定义的符号地址	
				addressing=advisit;
				return vp[i].pos;
			}
			i++;
		}

		if(ptrflag==1 ){	//已经使用了'@'符号，但是其后的标识符不是已经定义的符号地址
			if(*ptmp<'0' || *ptmp>'9'){	//未定义标识符不是数字，出错
				RES=51;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
					systemChecker.check(RES, sysLog);
				}
				return -1;
			}
			else{	//已修改。@7是非法的，认为7是立即数，前面不可出现其他符号 //未定义标识符是数字，其直接采用了间接寻址的方式，如@7 EA=（7）
				RES=56;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
					systemChecker.check(RES, sysLog);
				}
				return -1;			
			}
		}
		else{	//改为操作数为立即数 //操作数就是主存地址
			if(*ptmp=='+' || *ptmp=='-' || (*ptmp>='0' && *ptmp<='9')){
				if(advisit !=0){
					RES=56;
					if(systemChecker.showLevel(RES, sysLog)){
						printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
						systemChecker.check(RES, sysLog);
					}
					return -1;
				}
				advisit+=3;
				addressing=advisit;
				return atoi(ptmp);	
			}

			RES=52; //出现了未定义的符号，表达式非法
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d row:%d, operand:\'%s\' ",ID,line,op);
				systemChecker.check(RES, sysLog);
			}
			return -1; 
		}
	}

}

int editor::editorFromFile(FILE *pfile, assembler &A, int &dataNumInMData){	//editor in line ,dataNumInMData is num of data in MData
	int o,d,s;	//opcode, destination operand and source oprand
	//char line[5];
	char op[20];
	char od[20];
	char os[20];
	char comment[80]; //注释
	int opcodeType=0; //with 0,1 or 2 operand
	int addressing=0; //address visit way
	int addressingUnion;	//used to handle double operand addressing
	int i=0,j=0;	//i 记录实际行号，j 记录显示行号
	int dataCounter=0;	//data dim counter. 记录变量在数据段中的位置（从0开始）
	int state=0; //记录汇编情况，0位正常，其余为不正常。

	int v=0;	//the var num
	int tagv=0;	//标号项序号
	int jmpv=0;	//跳转项序号

	varNote varTable[20];	//max numbers of variables :20
	varNote tagTable[20];	//tag table
	varNote2 jumpTable[20];	//table about jump instructions

	if(displayMode) printf("display your statememts\n");
	if(displayMode) printf("%-4d",j);

	//fscanf(pfile,"%s%s",line,op);	//读取操作码
	fscanf(pfile,"%s",op);	//读取操作码，取消行号
	o=A.trans(op,j);  //need to change; trans char to int opcode
	opcodeType=o%10; //操作数个数，注释的opcodeType=3

	switch(opcodeType){
		case 0: d=s=0;
			if(displayMode) printf("%s\n", op); //无操作数
			break;
		case 1: s=0;
			fscanf(pfile,"%s",od);	//单操作数，读取目的操作数
			if(displayMode) printf("%s %s\n",op,od);
			break;
		case 2: 
			if(o!=910002){	//对于dim语句，特殊处理
				fscanf(pfile,"%s%s",od,os);	//双操作数，读取目的操作数和源操作数
				if(displayMode) printf("%s %s %s\n",op,od,os);	//故意少输出一个换行符，用于方面变量定义语句(修改)
			}
			else{
				fscanf(pfile,"%s%s",od,os);	//双操作数，读取目的操作数和源操作数
				if(displayMode) printf("%s %s %s ",op,od,os);	//故意少输出一个换行符，用于方面变量定义语句
			}
			break;
//		case 3:
//			fscanf(pfile,"%s",comment);
//			if(displayMode) printf("%s %s\n",op,comment);	//comment, do nothing
//			break;
		default: ;
	}

	if(o!=910002 && o!=920001 && o!=3 && opcodeType != 0){//not loc statement and not comment. 计算寻址方式
		if(opcodeType==1){
			
			if(o!=900421 && o!=900521 && o!=900621 && o!=900721 && o!=900821 && o!=904121){	//对于跳转语句和call，采用不同的出来操作数的方式
				d=A.trans2(od,varTable,j,addressing);	//trans char to int operand
				addressingUnion=addressing;
			}
			//else
			//	addressingUnion=3; //对于跳转语句，标号相当于立即数
		}
		if(opcodeType==2 ){	//双操作数，计算复合寻址方式
			d=A.trans2(od,varTable,j,addressing);	//trans char to int operand
			if(addressing==3){
				RES=197;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d operand:\'%s\' ",ID,od);
					systemChecker.check(RES, sysLog);
				}
				state=8;
			}
			addressingUnion=addressing;
			s=A.trans2(os,varTable,j,addressing);
			addressingUnion+=addressing*1000;
		}
	}


	while(o!=900000 && o!=-1 && i<codeBufferSize){	//900000:end, -1:syntax error
		if(o==920001){//标号
			//putchar('\n');
			int m=0;
			while(tagTable[m].valid!=0){ //保存标号
				if(!strcmp(od,tagTable[m].var)){ //标号重复
					RES=193;
					if(systemChecker.showLevel(RES, sysLog)){
						printf("D%-5d repeated tag:%s ",ID,od);
						systemChecker.check(RES, sysLog);
					}
					state=1;
					break;
				}
				m++;
			}
			
			//将标号信息记入标号表
			strcpy(tagTable[tagv].var,od);	//note the var name
			tagTable[tagv].pos=i;	//note the var pos
			tagTable[tagv].valid=1;		//valid==1 means the position in varTable has been used
			tagv++;			

		}
		else
		if(o==900421 || o==900521 || o==900621 || o==900721 || o==900821 || o==904121){	//将跳转信息记入跳转表
			strcpy(jumpTable[jmpv].var,od);
			jumpTable[jmpv].pos=i;
			jumpTable[jmpv].valid=1;
			jmpv++;
			
			//putchar('\n');
			codeBuffer[i]=o;	//write to codeBuffer
			codeBuffer[i+1]=3;	//visit address way, immediate 3
			codeBuffer[i+3]=0;	//没有源操作数
			//pptr->MCode.write(i,o);	//write to MCode
			//pptr->MCode.write(i+1,3);	//visit address way, immediate 3
			//pptr->MCode.write(i+2,d);
			//pptr->MCode.write(i+3,s);	
			i+=4;
			j++;	//LOC statement not increase the num of line
		}
		else
		if(o==910002){ //note new var
			if((od[0] >='a' && od[0] <='z') || (od[0] >='A' && od[0] <='Z')){	//var name must begin with letter
			
				int t=0;
				while(varTable[t].valid!=0){
					if(!strcmp(od,varTable[t].var)){ //变量重复
						RES=196;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d repeated indentifier:%s ",ID,od);
							systemChecker.check(RES, sysLog);
						}
						state=2;
						break;
					}
					t++;
				}

				strcpy(varTable[v].var,od);	//note the var name
				//varTable[v].pos=atoi(os);	//note the var pos
				varTable[v].pos=dataCounter;	//note the var pos
				varTable[v].valid=1;		//valid==1 means the position in varTable has been used
				v++;
				
				
				int len;	//声明的数组的长度
				int dataNum;	//赋初值的数组元素的个数+当前数据段中的偏移量
				char dataValue[40];	//初值字符串
				char fc;	//用于读取文件中初值的字符
				int i=0;
				char *cp;	//初值字符串指针
				int cflag=0;	//表示一个初值是否结束，以逗号分隔
				int k;	//将初值字符串中的初值转换成整数
				
				if(os[0]=='D'){	//表示定义的是数组
					len=atoi(os+1);
					dataNum=dataCounter;
					dataCounter+=len;
				

					fc=fgetc(pfile);

					while(fc!=';'){	//初值字符串以分号结束
						dataValue[i]=fc;	//将从文件中读取的字符保存到数组
						fc=fgetc(pfile);	//读下一个字符
						i++;
					}
					dataValue[i]=';';	//在字符串末尾加上分号，表示初值读完
					dataValue[i+1]='\0';	//表示字符串结束
					
					cp=dataValue;	//将cp指向初值字符串的第一个字符

					while(*cp!=';'){//初值字符串以分号结束
						if(((*cp>='0' && *cp <='9' ) || *cp=='-' || *cp=='+' || *cp=='\?' ) && cflag==0 ){
							if(dataNum+1>dataCounter){	//已经读取的初值数超过了申明时定义的数组长度，报错
								putchar('\n');
								RES=191;
								if(systemChecker.showLevel(RES, sysLog)){
									printf("D%-5d ",ID);
									systemChecker.check(RES, sysLog);
								}
								state=3;
								break;
							}	
							k=atoi(cp);	//将初值字符串中的初值转换成整数
							//pptr->MData.write(dataNum,k); //将初值保存到数据段
							dataBuffer[dataNum]=k;
							dataNum++;	//读取的字符数加1
							//printf("%d",k);
							cflag=1;
						}
						else
							if(*cp==','){	//读到了分隔字符逗号
								cflag=0;	//表示一个初值读完
							}

						cp++;	//指向初值字符串中的下一个字符
					}
					
					printf("%s\n",dataValue);
					dataNum=0;	//全部读完后读取的初值数归0
				}
				else{
					putchar('\n');
					if((os[0]>='0' && os[0] <='9' ) || os[0]=='-' || os[0]=='+' || os[0]=='\?'){	//声明的是一个字符地址（一个变量）,问号为初值不确定
						int k;
						k=atoi(os);	//读取变量的值
						//pptr->MData.write(dataCounter,k); //将变量写道数据段
						dataBuffer[dataCounter]=k;
						dataCounter++;
					}
					else{	//格式非法
						RES=192;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d ",ID);
							systemChecker.check(RES, sysLog);
						}
						state=4;
					}
				}
	

			}
			else{	//非法符号地址
				//putchar('\n');
				RES=13;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d variable name:\"%s\" ",ID,od);	//var name must begin with letter
					systemChecker.check(RES, sysLog);
				}
				state=5;
			}
		}//if(o==910002) end

		else{	//loc statement with line num 0
			
			if(o==3){
				fscanf(pfile,"%s",comment);
				if(displayMode) printf("%s %s\n",op,comment);	//comment, do nothing
			}
			else{
				//putchar('\n');
				//pptr->MCode.write(i,o);	//write to MCode
				
				if(o==800000) addressingUnion=d=s=0;	//对于停机语句要特殊处理，设置停机语句是为了方便汇编是对子程序的处理

				codeBuffer[i]=o;
				//pptr->MCode.write(i+1,addressingUnion);	//visit address way
				codeBuffer[i+1]=addressingUnion;
				//pptr->MCode.write(i+2,d);
				codeBuffer[i+2]=d;
				//pptr->MCode.write(i+3,s);
				codeBuffer[i+3]=s;

				i+=4;
				j++;	//LOC statement not increase the num of line
			}
		}

		if(displayMode) printf("%-4d",j);	//num of line
		
		fscanf(pfile,"%s",op);
		o=A.trans(op,j);  //need to change; trans char to int opcode
		opcodeType=o%10;

		switch(opcodeType){
			case 0: d=s=0;
				if(displayMode) printf("%s\n", op); //无操作数
				break;
			case 1: s=0;
				fscanf(pfile,"%s",od);	//单操作数，读取目的操作数
				if(displayMode) printf("%s %s\n",op,od);
				break;
			case 2: 
				if(o!=910002){	//对于dim语句，特殊处理
					fscanf(pfile,"%s%s",od,os);	//双操作数，读取目的操作数和源操作数
					if(displayMode) printf("%s %s %s\n",op,od,os);	//故意少输出一个换行符，用于方面变量定义语句(修改)
				}
				else{
					fscanf(pfile,"%s%s",od,os);	//双操作数，读取目的操作数和源操作数
					if(displayMode) printf("%s %s %s ",op,od,os);	//故意少输出一个换行符，用于方面变量定义语句
				}
				break;
//			case 3:
//				fscanf(pfile,"%s",comment);
//				if(displayMode) printf("%s\n",comment);	//comment, do nothing
//				break;
			default: ;
		}

		if(o!=910002 && o!=920001 && o!=3 && opcodeType != 0){//not loc statement and not comment 
			
			if(opcodeType==1){
				if(o!=900421 && o!=900521 && o!=900621 && o!=900721 && o!=900821 && o!=904121){	//对于跳转语句和call，采用不同的出来操作数的方式
					d=A.trans2(od,varTable,j,addressing);	//trans char to int operand
					addressingUnion=addressing;
				}
			}

			if(opcodeType==2 ){	
				d=A.trans2(od,varTable,j,addressing);	//trans char to int operand
				if(addressing==3){
					RES=197;
					if(systemChecker.showLevel(RES, sysLog)){
						printf("D%-5d operand:\'%s\' ",ID,od);
						systemChecker.check(RES, sysLog);
					}
					state=8;
				}
				addressingUnion=addressing;
				s=A.trans2(os,varTable,j,addressing);
				addressingUnion+=addressing*1000;
			}
		}

			
		//fscanf(pfile,"%s%s%s%s",line,op,od,os);
		//if(displayMode) printf("%s %s %s",op,od,os);
		//getchar();

		//o=A.trans(op,j);

		//opcodeType=o%10;

		//if(o!=910002 && o!=1 && opcodeType!=0 ){
		//	d=A.trans2(od,varTable,j,addressing);
		//	addressingUnion=addressing;
		//	s=A.trans2(os,varTable,j,addressing);
		//	addressingUnion+=addressing*1000;

		//	printf("%d",addressingUnion);
		//}		
	}//while(o!=900000 && o!=-1 && i<pptr->MCode.mSize) end

	//putchar('\n');	//补上一个换行符

	if(o==900000 && state==0){	//normal end
		//下面修改跳转标识
		int m=0;
		int n=0;
		while(tagTable[m].valid!=0){ //对照标号表和跳转表，把跳转位置写入代码段， 详见文档《标号与跳转》
			n=0;
			while(jumpTable[n].valid!=0){
				if(jumpTable[n].found==0){
					if(!strcmp(jumpTable[n].var,tagTable[m].var)){
						//pptr->MCode.write(jumpTable[n].pos+2,tagTable[m].pos);
						codeBuffer[jumpTable[n].pos+2]=tagTable[m].pos;
						jumpTable[n].found=1;	//表示该跳转位置写入
					}	
				}
				n++;
			}
			m++;
		}

		m=0;
		while(jumpTable[m].valid!=0){	//查看时候有没有对应标号的跳转项
			if(jumpTable[m].found==0){
				RES=194;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d unfound tag:%s ",ID,jumpTable[m].var);
					systemChecker.check(RES, sysLog);
				}
				state=7;
				break;
			}
			m++;
		}

		if(displayMode) printf("end edit\n");
		//pptr->MCode.write(i,o);
		codeBuffer[i]=o;

		i++;
		dataNumInMData=dataCounter; //数据段中数据的个数
		return i;

	}
	else{	//state!=0, abend
		if(o==-1){	
			
			RES=14;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d opcode:\"%s\" ",ID,op);
				systemChecker.check(RES, sysLog);
			}
		}
		else
			if(state==0){	//内容太多,超过了代码缓冲区的大小			
				RES=15;
				if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d numbers of lines:%d ",ID,i/4);
				systemChecker.check(RES, sysLog);
			}
		}
		return -1;
	}
}


int editor::ASM(FILE *pfile, assembler &A, char *s){
	int r;
	int dataInMData;	//数据段数据的数目
	r=editorFromFile(pfile,A,dataInMData);
	if(r==-1) return -1; //editor abend
	else{
		//saveToFile(pptr,r,dataInMData);
		FILE *fsave;		
		char tmps[50]; //用了保存新的文件名
		char *p;
		char ext[10]=".co"; //字符码文件的扩展名为co

		strcpy(tmps,s);
		p=tmps;

		while(*p!='.') p++;	//将字符码文件的扩展名改为xo
		*p='\0';
		strcat(tmps,ext);
		
		if((fsave=fopen(tmps, "w"))==0){	//文件打开失败
			RES=195;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d ",ID);
				systemChecker.check(RES, sysLog);
			}		
		}

		//saveMem(fsave, r);
		//saveMem(fsave, dataInMData);
		
		int i;
		for(i=0; i<r; i++){
			fprintf(fsave,"%d ",codeBuffer[i]);
		}

		fprintf(fsave,"%d ",dataInMData); //把数据长度写进文件，方便读写

		for(i=0; i<dataInMData ; i++){
			fprintf(fsave,"%d ",dataBuffer[i]);
		}
			
		fclose(fsave);
		return 0;
	}		
}

int editor::editorDataFromFile(FILE *pfile, memory &M, assembler &A){	//edit the content of memory at one time
	char line[5]; //row number
	char r[10];
	int i=0;

	if(displayMode) printf("enter the content in this memory ID:%d\n",M.ID);
	if(displayMode) printf("%-4d",i);	//num of line
	fscanf(pfile,"%s%s",line,r);
	if(displayMode) printf("%s\n", r);
	//getchar();
	while((r[0]=='-' || (r[0]>='0' && r[0]<='9')) && i<M.mSize){	//end edit
		M.write(i,atoi(r));	//trans r to num
		i++;
		if(displayMode) printf("%-4d",i);
		fscanf(pfile,"%s%s",line,r);
		if(displayMode) printf("%s\n", r);
		//getchar();
	}
	if(!strcmp(r,"END") || !strcmp(r,"end")){
		if(displayMode) printf("end edit\n");
		return 0;
	}
	else{
		
		RES=12;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d data:\"%s\" ",ID,r);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
}