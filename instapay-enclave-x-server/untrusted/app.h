#ifndef _APP_H_
#define _APP_H_

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>

#include "sgx_error.h"       /* sgx_status_t */
#include "sgx_eid.h"     /* sgx_enclave_id_t */

#ifndef TRUE
# define TRUE 1
#endif

#ifndef FALSE
# define FALSE 0
#endif

# define TOKEN_FILENAME   "enclave.token"
# define ENCLAVE_FILENAME "enclave.signed.so"

extern sgx_enclave_id_t global_eid;    /* global enclave id */

#if defined(__cplusplus)
extern "C" {
#endif

enum cross_message_type {
    
    CROSS_ALL_PREPARE_REQ = 1,
    CROSS_PREPARE_REQ     = 2,
    CROSS_PREPARE_RES     = 3,
    CROSS_ALL_PREPARED    = 4,

    CROSS_ALL_COMMIT_REQ  = 5,
    CROSS_COMMIT_REQ      = 6,
    CROSS_COMMIT_RES      = 7,
    CROSS_ALL_COMMITTED   = 8,

    CROSS_ALL_CONFIRM_REQ = 9,
    CROSS_CONFIRM_REQ     = 10,

    CROSS_ALL_REFUND_REQ  = 11,
    CROSS_REFUND_REQ      = 12,
};

enum cross_payment_server {
	CHAIN1_SERVER = 13,
	CHAIN2_SERVER = 14,
	CHAIN3_SERVER = 15,
};

typedef struct participant {

   	unsigned char party[41];
       	unsigned int payment_size;
   	unsigned int channel_ids[2];
   	int payment_amount[2];
} Participant;

typedef struct _cross_message {
    /********* common *********/
    cross_message_type type;

    /*** cross-payment ***/
    cross_payment_server server;

    /***** direct payment *****/
    unsigned int channel_id;
    int amount;
    //unsigned int counter;

    /*** multi-hop payment ****/
    
    unsigned int payment_num;
        /*
    unsigned int payment_size;
    unsigned int channel_ids[2];
    int payment_amount[2];
    */
    Participant participant[3];
    unsigned int e;

    /*** cross-payment ***/
    //cross_payment_server server;

} cross_message;

typedef struct _message {
    /********* common *********/
    unsigned int type;

    /***** direct payment *****/
    unsigned int channel_id;
    int amount;
    unsigned int counter;

    /*** multi-hop payment ****/
    unsigned int payment_num;
    unsigned int payment_size;
    unsigned int channel_ids[2];
    int payment_amount[2];
    unsigned int e;
} message;

int initialize_enclave(void);
void print_error_message(sgx_status_t ret);

void ocall_print_string(const char *str);

unsigned int ecall_accept_request_w(unsigned char *sender, unsigned char *receiver, unsigned int amount);
void ecall_add_participant_w(unsigned int payment_num, unsigned char *addr);
void ecall_update_sentagr_list_w(unsigned int payment_num, unsigned char *addr);
void ecall_update_sentupt_list_w(unsigned int payment_num, unsigned char *addr);
int ecall_check_unanimity_w(unsigned int payment_num, int which_list);
void ecall_update_payment_status_to_success_w(unsigned int payment_num);

/** 서버의 agreement request 메시지와 서명을 생성
 *
 * Out:     original_msg:   생성된 메시지의 plain text 주소
 *          output:         생성된 메시지의 signature 주소
 * In:      payment_num:    서버가 생성한 payment instance 번호
 *          payment_size:   payment data 길이 (channel_ids의 길이가 2이면, 값은 2)
 *                          길이는 반드시 1 또는 2임
 *          channel_ids:    해당 payment instance와 관련된 클라이언트의 채널 id 집합
 *          amount:         channel_ids에 포함된 id들과 매칭되는 각 payment amount
 *                          만약, 채널 A에 3을 지불해야 하면 -3이 되며, 지불받으면 +3이 됨
 */
void ecall_create_ag_req_msg_w(unsigned int payment_num, unsigned int payment_size, unsigned int *channel_ids, int *amount, unsigned char **original_msg, unsigned char **output);

/** 서버의 update request 메시지와 서명을 생성
 *
 * Out:     original_msg:   생성된 메시지의 plain text 주소
 *          output:         생성된 메시지의 signature 주소
 * In:      payment_num:    서버가 생성한 payment instance 번호
 *          payment_size:   payment data 길이 (channel_ids의 길이가 2이면, 값은 2)
 *                          길이는 반드시 1 또는 2임
 *          channel_ids:    해당 payment instance와 관련된 클라이언트의 채널 id 집합
 *          amount:         channel_ids에 포함된 id들과 매칭되는 각 payment amount
 *                          만약, 채널 A에 3을 지불해야 하면 -3이 되며, 지불받으면 +3이 됨
 */
void ecall_create_ud_req_msg_w(unsigned int payment_num, unsigned int payment_size, unsigned int *channel_ids, int *amount, unsigned char **original_msg, unsigned char **output);

/** 서버의 payment confirm 메시지와 서명을 생성
 *
 * Out:     original_msg:   생성된 메시지의 plain text 주소
 *          output:         생성된 메시지의 signature 주소
 * In:      payment_num:    서버가 생성한 payment instance 번호
 */
void ecall_create_confirm_msg_w(unsigned int payment_num, unsigned char **original_msg, unsigned char **output);

/** 클라이언트가 보낸 agreement response의 메시지 서명을 검증
 *
 * Returns: 1 이면 검증 성공, 0 이면 검증 실패
 * In:      pubaddr:   검증하려는 클라이언트의 공개 주소
 *          res_msg:   클라이언트의 agreement response 메시지의 plain text 주소
 *          res_sig:   클라이언트의 agreement response 메시지의 signature 주소
 */
unsigned int ecall_verify_ag_res_msg_w(unsigned char *pubaddr, unsigned char *res_msg, unsigned char *res_sig);

/** 클라이언트가 보낸 agreement response의 메시지 서명을 검증
 *
 * Returns: 1 이면 검증 성공, 0 이면 검증 실패
 * In:      pubaddr:   검증하려는 클라이언트의 공개 주소
 *          res_msg:   클라이언트의 agreement response 메시지의 plain text 주소
 *          res_sig:   클라이언트의 agreement response 메시지의 signature 주소
 */
unsigned int ecall_verify_ud_res_msg_w(unsigned char *pubaddr, unsigned char *res_msg, unsigned char *res_sig);


/*
 *
 *
 * InstaPay 3.0
 */
unsigned int ecall_cross_accept_request_w(
		unsigned char *chain1Server, 
		unsigned char *chain1Sender, 
		unsigned char *chain1Receiver, 
		unsigned int chain1Amount, 
		unsigned char *chain2Server, 
		unsigned char *chain2Sender, 
		unsigned char *chain2Receiver, 
		unsigned int chain2Amount,
		unsigned char *chain3Server, 
		unsigned char *chain3Sender, 
		unsigned char *chain3Receiver, 
		unsigned int chain3Amount,
		unsigned int numOfParticipants);

void ecall_cross_add_participant_w(unsigned int payment_num, unsigned char *addr);
void ecall_cross_update_preparedServer_list_w(unsigned int payment_num, unsigned char *addr);
void ecall_cross_update_committedServer_list_w(unsigned int payment_num, unsigned char *addr);
unsigned int ecall_cross_check_prepared_unanimity_w(unsigned int payment_num, int which_list);
unsigned int ecall_cross_check_committed_unanimity_w(unsigned int payment_num, int which_list);


unsigned int ecall_cross_create_all_prepare_req_msg_w(unsigned int payment_num, unsigned char *sender, unsigned char *middleMan, unsigned char *receiver, unsigned int sender_payment_size, unsigned int *sender_channel_ids, unsigned int middleMan_payment_size, unsigned int *middleMan_channel_ids, unsigned int receiver_payment_size, unsigned int *receiver_channel_ids, int *sender_amount, int *middleMan_amount, int *receiver_amount, unsigned char **original_msg, unsigned char **output);

unsigned int ecall_cross_create_all_prepare_req_msg_w2(unsigned int payment_num, unsigned char *sender, unsigned char *middleMan, unsigned char *receiver, unsigned char *receiver2, unsigned char *receiver3, unsigned int sender_payment_size, unsigned int *sender_channel_ids, unsigned int middleMan_payment_size, unsigned int *middleMan_channel_ids, unsigned int receiver_payment_size, unsigned int *receiver_channel_ids, unsigned int receiver2_payment_size, unsigned int *receiver2_channel_ids, unsigned int receiver3_payment_size, unsigned int *receiver3_channel_ids, int *sender_amount, int *middleMan_amount, int *receiver_amount, unsigned char **original_msg, unsigned char **output);


unsigned int ecall_cross_verify_all_prepared_res_msg_w(unsigned char *res_msg, unsigned char *res_sig, unsigned char *address);

unsigned int ecall_cross_create_all_commit_req_msg_w(unsigned int payment_num, unsigned char *sender, unsigned char *middleMan, unsigned char *receiver, unsigned int sender_payment_size, unsigned int *sender_channel_ids, unsigned int middleMan_payment_size, unsigned int *middleMan_channel_ids, unsigned int receiver_payment_size, unsigned int *receiver_channel_ids, int *sender_amount, int *middleMan_amount, int *receiver_amount, unsigned char **original_msg, unsigned char **output);

unsigned int ecall_cross_create_all_commit_req_msg_w2(unsigned int payment_num, unsigned char *sender, unsigned char *middleMan, unsigned char *receiver, unsigned char *receiver2, unsigned char *receiver3, unsigned int sender_payment_size, unsigned int *sender_channel_ids, unsigned int middleMan_payment_size, unsigned int *middleMan_channel_ids, unsigned int receiver_payment_size, unsigned int *receiver_channel_ids, unsigned int receiver2_payment_size, unsigned int *receiver2_channel_ids, unsigned int receiver3_payment_size, unsigned int *receiver3_channel_ids, int *sender_amount, int *middleMan_amount, int *receiver_amount, unsigned char **original_msg, unsigned char **output);

unsigned int ecall_cross_verify_all_committed_res_msg_w(unsigned char *res_msg, unsigned char *res_sig, unsigned char *address);

unsigned int ecall_cross_create_all_confirm_req_msg_w(unsigned int payment_num, unsigned char *sender, unsigned char *middleMan, unsigned char *receiver, unsigned int sender_payment_size, unsigned int *sender_channel_ids, unsigned int middleMan_payment_size, unsigned int *middleMan_channel_ids, unsigned int receiver_payment_size, unsigned int *receiver_channel_ids, int *sender_amount, int *middleMan_amount, int *receiver_amount, unsigned char **original_msg, unsigned char **output);

unsigned int ecall_cross_create_all_confirm_req_msg_w2(unsigned int payment_num, unsigned char *sender, unsigned char *middleMan, unsigned char *receiver, unsigned char *receiver2, unsigned char *receiver3, unsigned int sender_payment_size, unsigned int *sender_channel_ids, unsigned int middleMan_payment_size, unsigned int *middleMan_channel_ids, unsigned int receiver_payment_size, unsigned int *receiver_channel_ids, unsigned int receiver2_payment_size, unsigned int *receiver2_channel_ids, unsigned int receiver3_payment_size, unsigned int *receiver3_channel_ids, int *sender_amount, int *middleMan_amount, int *receiver_amount, unsigned char **original_msg, unsigned char **output);

void ecall_initSecp256k1CTX_w();

void ecall_cross_create_all_refund_req_msg_w(unsigned int payment_num, unsigned char **original_msg, unsigned char **output);

#if defined(__cplusplus)
}
#endif

#endif /* !_APP_H_ */
