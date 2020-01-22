#ifndef CA_H
#define CA_H

typedef void* OPAQUE;



typedef struct {
	int i;
	int j;
	int k;
} CA;

typedef unsigned char	uint8;
typedef unsigned short	uint16;
typedef unsigned int	uint32;
typedef unsigned long long		uint64;

typedef struct {
	uint16	lid;
	uint32	qpn;
	uint32	psn;
	uint64	raddr;
	uint32	rkey;

	uint64	subnet_prefix;
	uint64	interface_id;
} IBVAddress;

typedef struct {
	uint64_t subnet_prefix;
	uint64_t interface_id;
} IBVGid;

extern void test_ibv_struct(IBVAddress*);

extern char *go_ibv_get_error(OPAQUE);
extern int go_ibv_toggle_use_event(OPAQUE);

extern void go_ibv_init();

extern void* go_ibv_alloc_buffer(OPAQUE,int);

extern void* get_buf(unsigned int);
extern int go_ibv_get_device_list_num(OPAQUE);
extern OPAQUE go_ibv_get_device_list();
extern int go_ibv_free_device_list(OPAQUE);
extern const char* go_ibv_get_device_name(OPAQUE, int);
extern int go_ibv_set_device_index(OPAQUE,int);
extern int go_ibv_open_device(OPAQUE);
extern int go_ibv_close_device(OPAQUE);
extern int go_ibv_create_comp_channel(OPAQUE);
extern int go_ibv_destroy_comp_channel(OPAQUE);
extern int go_ibv_alloc_pd(OPAQUE);
extern int go_ibv_dealloc_pd(OPAQUE);
extern int go_ibv_reg_mr(OPAQUE);
extern int go_ibv_dereg_mr(OPAQUE);

extern int go_ibv_create_cq(OPAQUE,int,int);
extern int go_ibv_simple_create_qp(OPAQUE);
extern int go_ibv_destroy_cq(OPAQUE);

extern int go_ibv_init_qp(OPAQUE,unsigned char); // port
extern int go_ibv_query_port(OPAQUE, unsigned char); // port

extern int go_ibv_get_ibv_address_non_rdma(OPAQUE, IBVAddress*);
extern int go_ibv_get_ibv_address_non_rdma_nogen(OPAQUE, IBVAddress*);
extern int go_ibv_get_ibv_address(OPAQUE, IBVAddress* );

extern int go_ibv_post_recv(OPAQUE, int);
extern int go_ibv_query_gid(OPAQUE, int, int, IBVGid *);

extern int go_ibv_modify_qp_remote_ep(OPAQUE, uint32, uint32, uint16);

extern int go_ibv_recv_req_notify_cq(OPAQUE);
extern int go_ibv_send_req_notify_cq(OPAQUE);
extern int go_ibv_req_notify_cq(OPAQUE);
extern int go_ibv_post_send(OPAQUE,int,int);

extern int go_ibv_poll_cq_event(OPAQUE, int);

extern int go_ibv_ibv_cleanup(OPAQUE);

#endif /* CA_H */



