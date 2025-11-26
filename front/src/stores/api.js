import { ref } from 'vue';
import { defineStore } from 'pinia';

export const useApiStore = defineStore('api', () => {
    // localhost????
    const api = ref('http://localhost:9000/api');// 后端容器地址（本地开发环境，生产环境请通过配置对话框修改）
    const ipfsServiceUrl = ref('http://47.113.204.64:5001'); // ipfs服务地址（本地开发环境，生产环境请通过配置对话框修改）
    const chainServiceUrl = ref(''); // 链服务地址（tencent-chainmaker）
    const contractName = ref('z_test_chain');// 合约名称


    function setApi(url) {
        api.value = url;
    }

    function setIpfsServiceUrl(url) {
        ipfsServiceUrl.value = url;
    }

    function setChainServiceUrl(url) {
        chainServiceUrl.value = url;
    }

    function setContractName(name) {
        contractName.value = name;
    }

    return { api, setApi, contractName, chainServiceUrl, ipfsServiceUrl, setIpfsServiceUrl, setContractName, setChainServiceUrl };
});
