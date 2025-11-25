import { ref } from 'vue';
import { defineStore } from 'pinia';

export const useApiStore = defineStore('api', () => {
    const api = ref('https://服务器:9000/api');// 后端容器地址
    const ipfsServiceUrl = ref('http://服务器IP:5001'); // ipfs服务地址
    const chainServiceUrl = ref(''); // 链服务地址（tencent-chainmaker）
    const contractName = ref('chainqa');// 合约名称


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
