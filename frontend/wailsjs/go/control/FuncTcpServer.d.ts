// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {types} from '../models';

export function AddTCPServer(arg1:types.Server):Promise<types.ConnectResult>;

export function CheckConnectionStatus():Promise<void>;

export function DeleteTCPServer(arg1:number):Promise<types.ConnectResult>;

export function DisconnectClient(arg1:number,arg2:number):Promise<types.ConnectResult>;

export function GetAllTCPServers():Promise<types.ConnectResult>;

export function GetTCPServerData(arg1:number):Promise<types.ConnectResult>;

export function GetTCPServerStatus(arg1:number):Promise<types.ConnectResult>;

export function SendMessage(arg1:number,arg2:number,arg3:string):Promise<types.ConnectResult>;

export function StartTCPServer(arg1:number):Promise<types.ConnectResult>;

export function StopTCPServer(arg1:number):Promise<types.ConnectResult>;

export function UpdateTCPServer(arg1:types.Server):Promise<types.ConnectResult>;
