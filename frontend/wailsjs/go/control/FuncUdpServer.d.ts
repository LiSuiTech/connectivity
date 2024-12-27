// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {types} from '../models';

export function AddUdpServer(arg1:types.Server):Promise<types.ConnectResult>;

export function DeleteUdpServer(arg1:number):Promise<types.ConnectResult>;

export function DisconnectClient(arg1:number,arg2:number):Promise<types.ConnectResult>;

export function GetAllUdpServers():Promise<types.ConnectResult>;

export function GetUdpServerData(arg1:number):Promise<types.ConnectResult>;

export function GetUdpServerStatus(arg1:number):Promise<types.ConnectResult>;

export function SendMessage(arg1:number,arg2:number,arg3:string):Promise<types.ConnectResult>;

export function StartUdpServer(arg1:number):Promise<types.ConnectResult>;

export function StopUdpServer(arg1:number):Promise<types.ConnectResult>;

export function UpdateUdpServer(arg1:types.Server):Promise<types.ConnectResult>;